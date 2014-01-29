package cluster

import (
	"fmt"
	"testing"
	"time"
)


func TestMain(t *testing.T) {
   //var myid int
   // parse argument flags and get this server's id into myid
   server1 := make_server(1, /* config file */"config.txt")
   server2 := make_server(2, /* config file */"config.txt")
   server3 := make_server(3, /* config file */"config.txt")
   server4 := make_server(4, /* config file */"config.txt")
   // the returned server object obeys the Server interface above.
   // Let each server broadcast a message
   server1.get_outbox() <- &Envelope{Pid: 4, Msg: "hello there"}
   //server2.get_outbox() <- &Envelope{Pid: BROADCAST, Msg: "hello there"}
   select {
       case envelope1 := <- server1.get_inbox():
           fmt.Printf("Received msg from %d: '%s'\n", envelope1.Pid, envelope1.Msg)
       case envelope2 := <- server2.get_inbox():
           fmt.Printf("Received msg from %d to 2: '%s'\n", envelope2.Pid, envelope2.Msg)
       case envelope3 := <- server3.get_inbox():
           fmt.Printf("Received msg from %d to 3: '%s'\n", envelope3.Pid, envelope3.Msg)
       case envelope4 := <- server4.get_inbox():
           fmt.Printf("Received msg from %d to 4: '%s'\n", envelope4.Pid, envelope4.Msg)        
       case <- time.After(10 * time.Second): 
           println("Waited and waited. Ab thak gaya\n")
   }
}
