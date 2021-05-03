package user

import (
	"crypto/sha256"
	// "encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/registry"
	pb "github.com/harlow/go-micro-services/services/user/proto"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	// "io/ioutil"
	"log"
	"net"
	// "os"
	"time"
)

const name = "srv-user"

// Server implements the user service
type Server struct {
	users map[string]string

	Tracer   opentracing.Tracer
	Registry *registry.Client
	Port     int
	IpAddr	 string
	MongoSession 	*mgo.Session
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	if s.users == nil {
		s.users = loadUsers(s.MongoSession)
	}

	srv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout: 120 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.Tracer),
		),
	)

	pb.RegisterUserServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// // register the service
	// jsonFile, err := os.Open("config.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer jsonFile.Close()

	// byteValue, _ := ioutil.ReadAll(jsonFile)

	// var result map[string]string
	// json.Unmarshal([]byte(byteValue), &result)

	err = s.Registry.Register(name, s.IpAddr, s.Port)
	if err != nil {
		return fmt.Errorf("failed register: %v", err)
	}

	return srv.Serve(lis)
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
	s.Registry.Deregister(name)
}

// CheckUser returns whether the username and password are correct.
func (s *Server) CheckUser(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)

	// fmt.Printf("CheckUser")

	sum := sha256.Sum256([]byte(req.Password))
	pass := fmt.Sprintf("%x", sum)

	session, err := mgo.Dial("mongodb-user")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("user-db").C("user")
	var users []User
	err2 := c.Find(&bson.M{"username": req.Username}).All(&users)
	res.Correct = false
	if err2 != nil {
		log.Fatal(err2)
	} else {
		for _, user := range users {
			res.Correct = pass == user.Password
		}
	}

	// res.Correct = false
	// if true_pass, found := s.users[req.Username]; found {
	//     res.Correct = pass == true_pass
	// }
	
	// res.Correct = user.Password == pass

	// fmt.Printf("CheckUser %d\n", res.Correct)

	return res, nil
}

// CheckUser returns whether the username and password are correct.
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResult, error) {

	res := new(pb.RegisterResult)
	res.Correct = false

	user_name := req.Username
	sum := sha256.Sum256([]byte(req.Password))
	pass := fmt.Sprintf("%x", sum)
	age := req.Age
	sex := req.Sex
	mail := req.Mail
	phone := req.Phone
	orderhistory := ""

	session, err := mgo.Dial("mongodb-user")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("user-db").C("user")

	count, err := c.Find(&bson.M{"username": user_name}).Count()
	if err != nil {
		log.Fatal(err)
	}
	if count == 0{
		err = c.Insert(&User{user_name, pass, age, sex, mail, phone, orderhistory})
		if err != nil {
			log.Fatal(err)
		} else {
			res.Correct = true
		}
	}

	fmt.Printf("Done insert users\n")

	return res, nil
}


// CheckUser returns whether the username and password are correct.
func (s *Server) Modify(ctx context.Context, req *pb.ModifyRequest) (*pb.ModifyResult, error) {

	res := new(pb.ModifyResult)
	res.Correct = false

	user_name := req.Username
	sum := sha256.Sum256([]byte(req.Password))
	pass := fmt.Sprintf("%x", sum)
	age := req.Age
	sex := req.Sex
	mail := req.Mail
	phone := req.Phone
	orderhistory := ""

	session, err := mgo.Dial("mongodb-user")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("user-db").C("user")

	count, err := c.Find(&bson.M{"username": user_name}).Count()
	if err != nil {
		log.Fatal(err)
	}
	if count == 1{
		err := c.Remove(&bson.M{"username": user_name})
		if err != nil {
			log.Fatal(err)
		} else {
			err_2 := c.Insert(&User{user_name, pass, age, sex, mail, phone, orderhistory})
			if err_2 != nil {
				log.Fatal(err_2)
			} else {
				res.Correct = true
			}
		}
		
	}

	fmt.Printf("Done modify users\n")

	return res, nil
}

// CheckUser returns whether the username and password are correct.
func (s *Server) Delete(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	res.Correct = false

	// fmt.Printf("CheckUser")

	// sum := sha256.Sum256([]byte(req.Password))
	// pass := fmt.Sprintf("%x", sum)

	session, err := mgo.Dial("mongodb-user")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("user-db").C("user")
	count, err := c.Find(&bson.M{"username": req.Username}).Count()

	res.Correct = false
	if err != nil {
		log.Fatal(err)
	} 
	if count != 0 {
		
		err := c.Remove(&bson.M{"username": req.Username})
		if err != nil {
			log.Fatal(err)
		} else {
			res.Correct = true
		}
	}

	return res, nil
}

// CheckUser returns whether the username and password are correct.
func (s *Server) OrderHistoryUpdate(ctx context.Context, req *pb.OrderHistoryRequest) (*pb.OrderHistoryResult, error) {

	res := new(pb.OrderHistoryResult)
	res.Correct = false

	user_name := req.Username
	orderhistory := req.Orderhistory

	session, err := mgo.Dial("mongodb-user")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("user-db").C("user")

	var user_prof User
	err = c.Find(bson.M{"username": user_name}).One(&user_prof)

	if err != nil {
		log.Println("Failed get user data: ", err)
	} else {
		user_prof.Orderhistory = user_prof.Orderhistory + "; " + orderhistory
		err2 := c.Update(
			bson.M{"username": user_name},
			bson.M{"$set": bson.M{
				"orderhistory": user_prof.Orderhistory,
			}},
		)
		if err2 != nil {
			log.Println("Failed update user data: ", err2)
		} else {
			res.Correct = true
		}
	}

	fmt.Printf("Done update users orderhistory\n")

	return res, nil
}


// loadUsers loads hotel users from mongodb.
func loadUsers(session *mgo.Session) map[string]string {
	// session, err := mgo.Dial("mongodb-user")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()
	s := session.Copy()
	defer s.Close()
	c := s.DB("user-db").C("user")

	// unmarshal json profiles
	var users []User
	err := c.Find(bson.M{}).All(&users)
	if err != nil {
		log.Println("Failed get users data: ", err)
	}

	res := make(map[string]string)
	for _, user := range users {
		res[user.Username] = user.Password
	}

	fmt.Printf("Done load users\n")

	return res
}

// // insertUsers loads hotel users from mongodb.
// func (s *Server) insertUsers(session *mgo.Session, req *pb.RegisterRequest) (*pb.RegisterResult, error) {


// 	res := new(pb.RegisterResult)
// 	res.Correct = false

// 	user_name := req.Username
// 	sum := sha256.Sum256([]byte(req.Password))
// 	pass := fmt.Sprintf("%x", sum)
// 	age := req.Age
// 	sex := req.Sex
// 	mail := req.Mail
// 	phone := req.Phone
// 	orderhistory := ""

	
// 	s := session.Copy()
// 	defer s.Close()
// 	c := s.DB("user-db").C("user")

// 	// unmarshal json profiles
// 	// var users []User
// 	// err := c.Find(bson.M{}).All(&users)
// 	// if err != nil {
// 	// 	log.Println("Failed get users data: ", err)
// 	// }

// 	count, err := c.Find(&bson.M{"username": user_name}).Count()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if count == 0{
// 		err = c.Insert(&User{user_name, pass, age, sex, mail, phone, orderhistory})
// 		if err != nil {
// 			log.Fatal(err)
// 		} else {
// 			res.Correct = true
// 		}
// 	}

// 	fmt.Printf("Done insert users\n")

// 	return res, nil
// }

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
	Age int32 `bson:"age"`
	Sex string `bson:"sex"`
	Mail string `bson:"mail"`
	Phone string `bson:"phone"`
	Orderhistory string `bson:"orderhistory"`
}