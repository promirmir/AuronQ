//go:build windows || linux

package main

import (
 "errors"
 "testing"
 "time"
)

func TestThermalRecoveryNearTarget(t *testing.T) {
 c := newThermalController(0, 70, 81, 60)
 c.readTemperature = func(int)(int,error){return 69,nil}
 batch:=17
 for i:=0; i<9; i++ {
  c.lastCheck=time.Time{}
  next,_,_,action,err:=c.Adjust(batch)
  if err!=nil || next!=17 || action!="" {t.Fatalf("sample %d: batch=%d action=%q err=%v",i,next,action,err)}
 }
 c.lastCheck=time.Time{}
 next,_,pause,action,err:=c.Adjust(batch)
 if err!=nil || next!=18 || action!="increase" || pause!=0 {t.Fatalf("sample 10: batch=%d pause=%s action=%q err=%v",next,pause,action,err)}
}

func TestThermalSensorFailureStops(t *testing.T) {
 cases:=[]struct{name string; temp int; err error}{
  {"missing",-1,nil}, {"unphysical",130,nil},{"driver_error",0,errors.New("telemetry down")},{"hard_limit",81,nil},
 }
 for _,tt:=range cases {
  t.Run(tt.name,func(t *testing.T){
   c:=newThermalController(0,70,81,60)
   c.readTemperature=func(int)(int,error){return tt.temp,tt.err}
   _,_,_,action,err:=c.Adjust(40)
   if action!="stop"||err==nil {t.Fatalf("expected fail-closed stop, action=%q err=%v",action,err)}
  })
 }
}

func TestThermalReducesAtTarget(t *testing.T){
 c:=newThermalController(0,70,81,60)
 c.readTemperature=func(int)(int,error){return 70,nil}
 batch,_,pause,action,err:=c.Adjust(18)
 if err!=nil || action!="trim" || batch>=18 || pause<=0 {t.Fatalf("batch=%d pause=%s action=%q err=%v",batch,pause,action,err)}
}
