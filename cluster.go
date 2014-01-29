package cluster
import (
	"os"
	"strings"
	"bufio"
	"fmt"
	"net"
	"log"
	"encoding/json"
	"io"
	"strconv"
)


const (BROADCAST = -1)
type Envelope struct {
    // On the sender side, Pid identifies the receiving peer. If instead, Pid is
    // set to cluster.BROADCAST, the message is sent to all peers. On the receiver side, the
    // Id is always set to the original sender. If the Id is not found, the message is silently dropped
    Pid int
   
   
    // An id that globally and uniquely identifies the message, meant for duplicate detection at
    // higher levels. It is opaque to this package.
    MsgId int64
   
    // the actual message.
    Msg interface{}
}
   
type Server interface {
    // Id of this server
    get_pid() int
   
    // array of other servers' ids in the same cluster
    get_peers() []int
   
   
    // the channel to use to send messages to other peers
    // Note that there are no guarantees of message delivery, and messages
    // are silently dropped
    get_outbox() chan *Envelope
   
   
    // the channel to receive messages from other peers.
    get_inbox() chan *Envelope
}

type Node struct {
	pid int
	peers [] int
	address []string
	outbox chan*Envelope
	inbox chan*Envelope
}

func(n Node) get_pid() int{
	return n.pid
}
func(n Node) get_peers() []int{
	return n.peers
}
func(n Node) get_outbox() chan*Envelope{
	return n.outbox
}
func(n Node) get_inbox() chan*Envelope{
	return n.inbox
}

func make_server(server_id int, peer_file string) Server{
	var idList [] int;
	var addressList [] string;
	var curr_id int;
	var curr_address string;
	var tokens []string;
	var host string;
	var idNotFound int;
	/*Reading peer_file*/
	peer_file_stream,error := os.Open(peer_file)
	if error != nil {
		log.Fatal(error)
	}
	scanner := bufio.NewScanner(peer_file_stream);
	idNotFound=0;
	for scanner.Scan() {
		tokens = strings.Split(scanner.Text()," ");
		curr_id,error = strconv.Atoi(tokens[0]);
		if server_id == curr_id {
			host = curr_address;
			idNotFound = 1;
		}else {
			idList = append(idList,curr_id);
			curr_address = tokens[1];
			addressList = append(addressList,curr_address);
		}
	}
	if idNotFound == 0 {
		log.Fatal("error");
	}
	new_server := Node{server_id, idList, addressList, make(chan *Envelope), make(chan *Envelope)}	
	
	go func(){
		var buffer = make([]byte, 1024);
		for {
			listen, error := net.Listen("tcp", host);
			//defer listen.Close();
			if error != nil { 
		        	fmt.Printf("Error creating listener: %s\n", error ); 
		        	os.Exit(1); 
			}
			for{
				var readSuccess = true;
				con, error := listen.Accept();
					if error != nil {
						fmt.Printf("Error: Accepting data: %s\n", error); os.Exit(2); 
					}
				for readSuccess{
					size, error := con.Read(buffer);
					switch error {
						case io.EOF:
							readSuccess = false;
						case nil:
					        var env Envelope;
					        json.Unmarshal(buffer[0:size],&env);
					        new_server.get_inbox()<-&env;
						default:
		                	fmt.Printf("Error: Reading data : %s \n", error);
		                	readSuccess = false;
					}
				}
			//con.Close();
		   }
		}
	}()
	
	go func(){
		for {	
			sendEnv := <- new_server.get_outbox();
			servid := sendEnv.Pid;
			sendEnv.Pid = server_id;
			size,_:=json.Marshal(sendEnv);
			var listen_host string;
			if servid !=-1	{
				for key := range(idList){
					if idList[key] == servid {
						listen_host = addressList[key]
						fmt.Println(listen_host);
						break
					}
				}
				con, error := net.Dial("tcp", listen_host);
				if error != nil {
					fmt.Printf("Host not found: '%s'\n", error ); 
					os.Exit(1); 
				}
				in, error := con.Write(size);
				if error != nil { fmt.Printf("Error sending data: %s, in: %d\n", error, in ); os.Exit(2); }
			}else {
				for key := range(idList){
					if idList[key] != server_id {
						listen_host = addressList[key];
						fmt.Printf("host is  %s\n", listen_host);
						con, error := net.Dial("tcp", listen_host);
						//defer con.Close();
						if error != nil { fmt.Printf("Host not found1: %s\n", error ); os.Exit(1); }
						in, error := con.Write(size);
						if error != nil { fmt.Printf("Error sending data: %s, in: %d\n", error, in ); os.Exit(2); }
					}
				}
			}
		}
	}()
	return new_server
}


