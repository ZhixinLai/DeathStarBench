package admin

import (

	// "encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/registry"
	pb "github.com/harlow/go-micro-services/services/admin/proto"

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

const name = "srv-admin"

type Server struct {
	Tracer       opentracing.Tracer
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
	Registry     *registry.Client
}

func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
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
	pb.RegisterAdminServer(srv, s)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
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

//Checker the password and email input to make sure they are matched with the data in the database
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	res := new(pb.LoginReply)
	session, err := mgo.Dial("mongodb-admin")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	c := session.DB("admin-db").C("admin")
	admin := new(Admin)
	err1 := c.Find(bson.M{"email": req.Email}).One(admin)
	res.Correct = false
	if err1 != nil {
		log.Fatal(err1)
	} else {
		if admin != nil {
			res.Correct = req.Password == admin.Password
		}
	}
	return res, nil
}
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	res := new(pb.RegisterReply)
	res.Correct = false
	name := req.Name
	email := req.Email
	password := req.Password
	hotels := req.Hotels
	id := req.Id

	session, err := mgo.Dial("mongodb-admin")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("admin-db").C("admin")
	count, err := c.Find(&bson.M{"name": name}).Count()
	if count == 0 {
		err = c.Insert(&Admin{name, email, password, hotels, id})
		if err != nil {
			log.Fatal(err)
		} else {
			res.Correct = true
		}
	}
	fmt.Printf("Done insert users\n")

	return res, nil
}

func (s *Server) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateReply, error) {
	res := new(pb.UpdateReply)
	res.Correct = false
	id := req.Id
	target := req.Target
	content := req.Content
	//get that the orignal content
	session, err := mgo.Dial("mongodb-profile")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	c := session.DB("profile-db").C("hotels")
	err = c.Update(bson.M{"id": id}, bson.M{"$set": bson.M{target: content}})
	if err == nil {
		res.Correct = true
	}
	return res, nil

}
func (s *Server) CheckHotel(ctx context.Context, req *pb.CheckRequest) (*pb.CheckReply, error) {
	res := new(pb.CheckReply)
	res.Correct = false
	id := req.Id
	session, err := mgo.Dial("mongodb-admin")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	c := session.DB("admin-db").C("admin")
	admin := new(Admin)
	err1 := c.Find(bson.M{"email": req.Email}).One(admin)
	if err1 == nil {
		for i := 0; i < len(admin.Hotels); i++ {
			if admin.Hotels[i] == id {
				res.Correct = true
			}
		}
	}
	return res, nil
}

type Admin struct {
	Name     string   `bason:"name"`
	Email    string   `bason:"email"`
	Password string   `bson:"password"`
	Hotels   []string `bason: "hotels"`
	Id       string   `bason: "id"`
}
