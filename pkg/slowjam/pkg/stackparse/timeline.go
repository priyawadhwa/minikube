package stackparse

import (
	"fmt"
	"strings"
	"time"

	"github.com/maruel/panicparse/stack"
)

// SuggestedIgnore are goroutines that we recommend ignoring
var SuggestedIgnore = []string{
	"signal.init.0",
	"trace.Start",
	"stacklog.Start",
	"glog.init.0",
	"klog.init.0",
	"localbinary.(*Plugin).AttachStream",
	"rpc.(*DefaultRPCClientDriverFactory).NewRPCClientDriver",
	"http.(*http2Transport).newClientConn",
}

type Timeline struct {
	Start      time.Time
	End        time.Time
	Samples    int
	Goroutines map[int]*GoroutineTimeline
}

type GoroutineTimeline struct {
	ID        int
	Signature stack.Signature
	Layers    []*Layer
}

type Layer struct {
	Calls []*Call
}

type Call struct {
	StartDelta time.Duration
	EndDelta   time.Duration
	lastSeen   time.Time
	Samples    int
	Args       stack.Args
	Name       string
	Package    string
}

// SimplifyTimeline flattens overlapping layers from call-stacks in a timeline
func SimplifyTimeline(tl *Timeline) *Timeline {
	fmt.Printf("SIMPLIFY\n")
	newGoroutines := map[int]*GoroutineTimeline{}
	for gid, g := range tl.Goroutines {
		fmt.Printf("simplifying %d ...\n", gid)
		newLayers := []*Layer{}
		for il, l := range g.Layers {
			newCalls := []*Call{}
			for _, c := range l.Calls {
				// If it's less than .25%, omit
				if c.Samples*250 < tl.Samples {
					fmt.Printf("%d: dropping %s due to sample size (%d, duration %s)\n", gid, c.Name, c.Samples, c.EndDelta-c.StartDelta)
					continue
				}
				if il > 0 && il != len(g.Layers)-1 {
					drop := false
					above := g.Layers[il-1]
					for _, oc := range above.Calls {
						if oc.StartDelta == c.StartDelta && oc.EndDelta == c.EndDelta && c.Package == oc.Package {
							fmt.Printf("%d: dropping due to overlap: %s\n", gid, c.Name)
							drop = true
							break
						}
					}
					if drop {
						continue
					}
				}
				newCalls = append(newCalls, c)
			}
			if len(newCalls) < 1 {
				fmt.Printf("%d: dropping layer with %d calls due to lack of interesting calls\n", gid, len(l.Calls))
				continue
			}
			newLayers = append(newLayers, &Layer{Calls: newCalls})
		}
		if len(newLayers) < 1 {
			fmt.Printf("%d: dropping goroutine %d due to lack of layers\n", g.ID)
			continue
		}
		newGoroutines[gid] = &GoroutineTimeline{g.ID, g.Signature, newLayers}
	}
	fmt.Printf("simplified from %d to %d goroutines\n", len(tl.Goroutines), len(newGoroutines))
	return &Timeline{
		Start:      tl.Start,
		End:        tl.End,
		Samples:    tl.Samples,
		Goroutines: newGoroutines,
	}
}

// CreateTimeline creates a timeline from stack samples
func CreateTimeline(samples []*StackSample, ignoreCreators []string) *Timeline {
	ig := map[string]bool{}
	for _, i := range ignoreCreators {
		ig[i] = true
	}

	tl := &Timeline{
		Start:      samples[0].Time,
		End:        samples[len(samples)-1].Time,
		Goroutines: map[int]*GoroutineTimeline{},
	}
	for _, s := range samples {
		tl.Samples++
		for _, g := range s.Context.Goroutines {
			if ig[g.CreatedBy.Func.PkgDotName()] {
				continue
			}
			if tl.Goroutines[g.ID] == nil {
				tl.Goroutines[g.ID] = &GoroutineTimeline{
					ID:        g.ID,
					Signature: g.Signature,
					Layers:    []*Layer{},
				}
			}
			for depth, c := range g.Signature.Stack.Calls {
				if internalCall(c) {
					continue
				}
				thisCall := &Call{
					StartDelta: s.Time.Sub(tl.Start),
					Name:       c.Func.PkgDotName(),
					Package:    c.Func.PkgName(),
					Args:       c.Args,
					lastSeen:   s.Time,
					Samples:    1,
				}

				level := len(g.Signature.Stack.Calls) - depth - 1
				// fmt.Printf("level=%d, depth=%d call=%+v\n", level, depth, thisCall)
				// New layer!
				missing := level - (len(tl.Goroutines[g.ID].Layers) - 1)
				// fmt.Printf("%d has %d layers: missing=%d\n", g.ID, len(tl.Goroutines[g.ID].Layers), missing)
				if missing > 0 {
					//	fmt.Printf("missing %d levels\n", missing)
					for i := 0; i < missing; i++ {
						tl.Goroutines[g.ID].Layers = append(tl.Goroutines[g.ID].Layers, &Layer{Calls: []*Call{}})
					}
					tl.Goroutines[g.ID].Layers[level].Calls = []*Call{thisCall}
					continue
				}
				// Existing layer
				calls := tl.Goroutines[g.ID].Layers[level].Calls
				if len(calls) == 0 {
					// 		fmt.Printf("new call on level %d: %s\n", level, thisCall.Name)
					tl.Goroutines[g.ID].Layers[level].Calls = []*Call{thisCall}
					continue
				}
				lc := calls[len(calls)-1]
				// Existing call with the same name or short sample size
				if lc.Name == c.Func.PkgDotName() && lc.EndDelta == 0 && (lc.Samples < 3 || SameArgs(lc.Args, c.Args)) {
					lc.Samples++
					lc.lastSeen = s.Time
					continue
				}
				// End the previous call & add a new one
				// Err on the smaller time-scale: was this a 1ms call or a 100ms call?
				lc.EndDelta = lc.lastSeen.Sub(tl.Start)
				tl.Goroutines[g.ID].Layers[level].Calls = append(tl.Goroutines[g.ID].Layers[level].Calls, thisCall)
			}
		}
	}

	// End any trailing calls
	for _, g := range tl.Goroutines {
		for _, l := range g.Layers {
			if len(l.Calls) == 0 {
				continue
			}
			lc := l.Calls[len(l.Calls)-1]
			if lc.EndDelta == 0 {
				lc.EndDelta = lc.lastSeen.Sub(tl.Start)
			}
		}
	}

	return tl
}

func internalCall(c stack.Call) bool {
	if c.Func.PkgName() == "syscall" {
		return true
	}
	if c.Func.IsExported() {
		return false
	}
	if c.IsStdlib || strings.Contains(c.SrcPath, "/go/src/") {
		return true
	}
	return false
}

// equal returns true only if both arguments are exactly equal.
func SameArgs(a stack.Args, b stack.Args) bool {
	if a.Elided != b.Elided {
		return false
	}
	if len(a.Values) != len(b.Values) {
		return false
	}
	for i, l := range a.Values {
		if l.Value != b.Values[i].Value {
			// the value is different, maybe the function is?
			return false
		}
	}
	return true
}
