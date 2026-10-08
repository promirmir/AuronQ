package main

import (
 "strings"
 "testing"
 "time"
)

func TestAdaptiveAgentThermalPriority(t *testing.T) {
 a:=newAdaptiveMiningAgent()
 start:=time.Unix(1000,0)
 a.Observe(start,17,17,200*time.Millisecond,65,70,60,0)
 got,reason:=a.Observe(start.Add(21*time.Second),17,17,6*time.Second,69,70,60,0)
 if got!=17 || reason!="" {t.Fatalf("hot GPU should not probe: %d %q",got,reason)}
 got,reason=a.Observe(start.Add(43*time.Second),17,17,6*time.Second,65,70,60,80*time.Millisecond)
 if got!=17 || reason!="" {t.Fatalf("GPU thermal cooldown must suppress probe: %d %q",got,reason)}
}

func TestAdaptiveAgentProbeAndRollback(t *testing.T) {
 a:=newAdaptiveMiningAgent()
 start:=time.Unix(1000,0)
 a.Observe(start,17,17,time.Second,65,70,60,0)
 batch,action:=a.Observe(start.Add(21*time.Second),17,120,6*time.Second,65,70,60,0)
 if batch!=18 || !strings.Contains(action,"probe") {t.Fatalf("expected probe: %d %q",batch,action)}
 batch,action=a.Observe(start.Add(42*time.Second),18,60,6*time.Second,65,70,60,0)
 if batch!=17 || !strings.Contains(action,"rollback") {t.Fatalf("expected rollback: %d %q",batch,action)}
}

func TestAdaptiveAgentNeverExceedsCap(t *testing.T) {
 a:=newAdaptiveMiningAgent()
 start:=time.Unix(1000,0)
 a.Observe(start,20,20,time.Second,64,70,20,0)
 got,_:=a.Observe(start.Add(30*time.Second),20,120,6*time.Second,64,70,20,0)
 if got!=20 {t.Fatalf("unexpected batch beyond cap: %d",got)}
}
