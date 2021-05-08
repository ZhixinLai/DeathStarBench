package frontend

import (
	"encoding/json"
	"fmt"
	"github.com/harlow/go-micro-services/dialer"
	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/admin/proto"
	"github.com/harlow/go-micro-services/services/profile/proto"
	"github.com/harlow/go-micro-services/services/recommendation/proto"
	"github.com/harlow/go-micro-services/services/reservation/proto"
	"github.com/harlow/go-micro-services/services/search/proto"
	"github.com/harlow/go-micro-services/services/user/proto"
	"github.com/harlow/go-micro-services/tracing"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"strconv"
)

// Server implements frontend service
type Server struct {
	searchClient         search.SearchClient
	profileClient        profile.ProfileClient
	recommendationClient recommendation.RecommendationClient
	userClient           user.UserClient
	adminClient          admin.AdminClient
	reservationClient    reservation.ReservationClient
	IpAddr               string
	Port                 int
	Tracer               opentracing.Tracer
	Registry             *registry.Client
}

// Run the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	if err := s.initSearchClient("srv-search"); err != nil {
		return err
	}

	if err := s.initProfileClient("srv-profile"); err != nil {
		return err
	}

	if err := s.initRecommendationClient("srv-recommendation"); err != nil {
		return err
	}

	if err := s.initUserClient("srv-user"); err != nil {
		return err
	}

	if err := s.initReservation("srv-reservation"); err != nil {
		return err
	}

	if err := s.initAdminClient("srv-admin"); err != nil {
		return err
	}
	// fmt.Printf("frontend before mux\n")

	mux := tracing.NewServeMux(s.Tracer)
	mux.Handle("/", http.FileServer(http.Dir("services/frontend/static")))
	mux.Handle("/hotels", http.HandlerFunc(s.searchHandler))
	mux.Handle("/recommendations", http.HandlerFunc(s.recommendHandler))
	mux.Handle("/user", http.HandlerFunc(s.userHandler))
	mux.Handle("/userregister", http.HandlerFunc(s.userRegisterHandler))
	mux.Handle("/usermodify", http.HandlerFunc(s.userModifyHandler))
	mux.Handle("/userdelete", http.HandlerFunc(s.userDeleteHandler))
	mux.Handle("/userevaluate", http.HandlerFunc(s.userEvaluateHandler))
	mux.Handle("/reservation", http.HandlerFunc(s.reservationHandler))
	mux.Handle("/cancelreservation", http.HandlerFunc(s.cancelReservationHandler))
	mux.Handle("/adminlogin", http.HandlerFunc(s.adminLoginHandler))
	mux.Handle("/daminregister", http.HandlerFunc(s.adminRegisterHandler))
	mux.Handle("/updateProfile", http.HandlerFunc(s.updateProfileHandler))
	// fmt.Printf("frontend starts serving\n")

	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux)
}
func (s *Server) initAdminClient(name string) error {
	conn, err := dialer.Dial(
		name,
		dialer.WithTracer(s.Tracer),
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.adminClient = admin.NewAdminClient(conn)
	return nil
}
func (s *Server) initSearchClient(name string) error {
	conn, err := dialer.Dial(
		name,
		dialer.WithTracer(s.Tracer),
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.searchClient = search.NewSearchClient(conn)
	return nil
}

func (s *Server) initProfileClient(name string) error {
	conn, err := dialer.Dial(
		name,
		dialer.WithTracer(s.Tracer),
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.profileClient = profile.NewProfileClient(conn)
	return nil
}

func (s *Server) initRecommendationClient(name string) error {
	conn, err := dialer.Dial(
		name,
		dialer.WithTracer(s.Tracer),
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.recommendationClient = recommendation.NewRecommendationClient(conn)
	return nil
}

func (s *Server) initUserClient(name string) error {
	conn, err := dialer.Dial(
		name,
		dialer.WithTracer(s.Tracer),
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.userClient = user.NewUserClient(conn)
	return nil
}

func (s *Server) initReservation(name string) error {
	conn, err := dialer.Dial(
		name,
		dialer.WithTracer(s.Tracer),
		dialer.WithBalancer(s.Registry.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.reservationClient = reservation.NewReservationClient(conn)
	return nil
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	// fmt.Printf("starts searchHandler\n")

	// in/out dates from query params
	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	// lan/lon from query params
	sLat, sLon := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if sLat == "" || sLon == "" {
		http.Error(w, "Please specify location params", http.StatusBadRequest)
		return
	}

	Lat, _ := strconv.ParseFloat(sLat, 32)
	lat := float32(Lat)
	Lon, _ := strconv.ParseFloat(sLon, 32)
	lon := float32(Lon)

	// fmt.Printf("starts searchHandler querying downstream\n")

	// search for best hotels
	searchResp, err := s.searchClient.Nearby(ctx, &search.NearbyRequest{
		Lat:     lat,
		Lon:     lon,
		InDate:  inDate,
		OutDate: outDate,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Printf("searchHandler gets searchResp\n")
	// for _, hid := range searchResp.HotelIds {
	// 	fmt.Printf("search Handler hotelId = %s\n", hid)
	// }

	// grab locale from query params or default to en
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	reservationResp, err := s.reservationClient.CheckAvailability(ctx, &reservation.Request{
		CustomerName: "",
		HotelId:      searchResp.HotelIds,
		InDate:       inDate,
		OutDate:      outDate,
		RoomNumber:   1,
	})

	// fmt.Printf("searchHandler gets reserveResp\n")
	// fmt.Printf("searchHandler gets reserveResp.HotelId = %s\n", reservationResp.HotelId)

	// hotel profiles
	profileResp, err := s.profileClient.GetProfiles(ctx, &profile.Request{
		HotelIds: reservationResp.HotelId,
		Locale:   locale,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Printf("searchHandler gets profileResp\n")

	json.NewEncoder(w).Encode(geoJSONResponse(profileResp.Hotels))
}

func (s *Server) recommendHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	sLat, sLon := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if sLat == "" || sLon == "" {
		http.Error(w, "Please specify location params", http.StatusBadRequest)
		return
	}
	Lat, _ := strconv.ParseFloat(sLat, 64)
	lat := float64(Lat)
	Lon, _ := strconv.ParseFloat(sLon, 64)
	lon := float64(Lon)

	require := r.URL.Query().Get("require")
	if require != "dis" && require != "rate" && require != "price" {
		http.Error(w, "Please specify require params", http.StatusBadRequest)
		return
	}

	// recommend hotels
	recResp, err := s.recommendationClient.GetRecommendations(ctx, &recommendation.Request{
		Require: require,
		Lat:     float64(lat),
		Lon:     float64(lon),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// grab locale from query params or default to en
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	// hotel profiles
	profileResp, err := s.profileClient.GetProfiles(ctx, &profile.Request{
		HotelIds: recResp.HotelIds,
		Locale:   locale,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(geoJSONResponse(profileResp.Hotels))
}
func (s *Server) adminRegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()
	// hotels := []string {"1", "2", "3"}
	hotels := r.URL.Query().Get("hotels")
	if hotels == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}
	var hotelsArr []string
	json.Unmarshal([]byte(hotels), hotelsArr)
	name, email, password, id := r.URL.Query().Get("name"), r.URL.Query().Get("email"), r.URL.Query().Get("password"), r.URL.Query().Get("id")
	if name == "" || email == "" || password == "" || id == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}
	recResp, err := s.adminClient.Register(ctx, &admin.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
		Hotels:   hotelsArr,
		Id:       id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	str := "Register successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your register input. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}
func (s *Server) updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()
	email, password := r.URL.Query().Get("password"), r.URL.Query().Get("password")
	if email == "" || password == "" {
		http.Error(w, "Please specify email /password params", http.StatusBadRequest)
		return
	}
	recResp, err := s.adminClient.Login(ctx, &admin.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if recResp.Correct == false {
		str := "Failed. Please check your username and password. "
		res := map[string]interface{}{
			"message": str,
		}

		json.NewEncoder(w).Encode(res)
		return
	} else {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Please specify id params", http.StatusBadRequest)
			return
		}
		checkResp, err1 := s.adminClient.CheckHotel(ctx, &admin.CheckRequest{
			Email: email,
			Id:    id,
		})
		if err1 != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if checkResp.Correct == false {
			str := "It is not your hotel, you could not update it "
			res := map[string]interface{}{
				"message": str,
			}

			json.NewEncoder(w).Encode(res)
			return
		} else {
			target, content := r.URL.Query().Get("target"), r.URL.Query().Get("content")
			if target == "" {
				http.Error(w, "Please specify target/content params", http.StatusBadRequest)
				return
			}
			updateResp, err2 := s.adminClient.Update(ctx, &admin.UpdateRequest{
				Id:      id,
				Target:  target,
				Content: content,
			})
			if err2 != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			str1 := "Update fail"
			if updateResp.Correct == true {
				str1 = "Success"
			}
			res := map[string]interface{}{
				"message": str1,
			}
			json.NewEncoder(w).Encode(res)
			return
		}
	}

}
func (s *Server) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()
	//get parameter
	email, password := r.URL.Query().Get("email"), r.URL.Query().Get("password")
	if email == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}
	recResp, err := s.adminClient.Login(ctx, &admin.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	str := "Login successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Login successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) userRegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password, age, sex, mail, phone := r.URL.Query().Get("username"), r.URL.Query().Get("password"), r.URL.Query().Get("age"), r.URL.Query().Get("sex"), r.URL.Query().Get("mail"), r.URL.Query().Get("phone")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	age_int, err := strconv.ParseInt(age, 10, 32)
	if err != nil {
		panic(err)
	}

	// Register
	recResp, err := s.userClient.Register(ctx, &user.RegisterRequest{
		Username: username,
		Password: password,
		Age:      int32(age_int),
		Sex:      sex,
		Mail:     mail,
		Phone:    phone,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Register successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) userModifyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password, age, sex, mail, phone := r.URL.Query().Get("username"), r.URL.Query().Get("password"), r.URL.Query().Get("age"), r.URL.Query().Get("sex"), r.URL.Query().Get("mail"), r.URL.Query().Get("phone")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	age_int, err := strconv.ParseInt(age, 10, 32)
	if err != nil {
		panic(err)
	}

	// Modify
	recResp, err := s.userClient.Modify(ctx, &user.ModifyRequest{
		Username: username,
		Password: password,
		Age:      int32(age_int),
		Sex:      sex,
		Mail:     mail,
		Phone:    phone,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Modify successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Delete user
	recResp, err := s.userClient.Delete(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Delete successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	}

	res := map[string]interface{}{
		"message": str,
	}

	json.NewEncoder(w).Encode(res)
}

func (s *Server) userEvaluateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	if !checkDataFormat(inDate) || !checkDataFormat(outDate) {
		http.Error(w, "Please check inDate/outDate format (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	customerName := r.URL.Query().Get("customerName")
	if customerName == "" {
		http.Error(w, "Please specify customerName params", http.StatusBadRequest)
		return
	}

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	str := "Score successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	} else {

		hotelId := r.URL.Query().Get("hotelId")
		score := r.URL.Query().Get("score")

		score_float, err := strconv.ParseFloat(score, 64)
		if err != nil {
			panic(err)
		}

		// update order history for user
		orderhistory := "hotelId: " + hotelId + ", inDate: " + inDate + ", outDate: " + outDate + ", score: " + score
		orderhistoryResp, err := s.userClient.OrderHistoryUpdate(ctx, &user.OrderHistoryRequest{
			Username:     username,
			Orderhistory: orderhistory,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if orderhistoryResp.Correct == false {
			str = "Failed. "
		}

		// update score in profile of hotel
		profileResp, err := s.profileClient.UpdateScore(ctx, &profile.ScoreRequest{
			HotelId: hotelId,
			Score:   float32(score_float),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if profileResp.Correct == false {
			str = "Failed. "
		}
	}

	res := map[string]interface{}{
		"message": str,
	}
	json.NewEncoder(w).Encode(res)
}

func (s *Server) reservationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	if !checkDataFormat(inDate) || !checkDataFormat(outDate) {
		http.Error(w, "Please check inDate/outDate format (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	hotelId := r.URL.Query().Get("hotelId")
	if hotelId == "" {
		http.Error(w, "Please specify hotelId params", http.StatusBadRequest)
		return
	}

	customerName := r.URL.Query().Get("customerName")
	if customerName == "" {
		http.Error(w, "Please specify customerName params", http.StatusBadRequest)
		return
	}

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	numberOfRoom := 0
	num := r.URL.Query().Get("number")
	if num != "" {
		numberOfRoom, _ = strconv.Atoi(num)
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Reserve successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	} else {
		// Make reservation
		resResp, err := s.reservationClient.MakeReservation(ctx, &reservation.Request{
			CustomerName: customerName,
			HotelId:      []string{hotelId},
			InDate:       inDate,
			OutDate:      outDate,
			RoomNumber:   int32(numberOfRoom),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(resResp.HotelId) == 0 {
			str = "Failed. Already reserved. "
		}
	}
	res := map[string]interface{}{
		"message": str,
	}
	json.NewEncoder(w).Encode(res)
}

func (s *Server) cancelReservationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	if !checkDataFormat(inDate) || !checkDataFormat(outDate) {
		http.Error(w, "Please check inDate/outDate format (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	hotelId := r.URL.Query().Get("hotelId")
	if hotelId == "" {
		http.Error(w, "Please specify hotelId params", http.StatusBadRequest)
		return
	}

	customerName := r.URL.Query().Get("customerName")
	if customerName == "" {
		http.Error(w, "Please specify customerName params", http.StatusBadRequest)
		return
	}

	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Please specify username and password", http.StatusBadRequest)
		return
	}

	numberOfRoom := 0
	num := r.URL.Query().Get("number")
	if num != "" {
		numberOfRoom, _ = strconv.Atoi(num)
	}

	// Check username and password
	recResp, err := s.userClient.CheckUser(ctx, &user.Request{
		Username: username,
		Password: password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str := "Cancel successfully!"
	if recResp.Correct == false {
		str = "Failed. Please check your username and password. "
	} else {

		// Cancel reservation
		resResp, err := s.reservationClient.CancelReservation(ctx, &reservation.Request{
			CustomerName: customerName,
			HotelId:      []string{hotelId},
			InDate:       inDate,
			OutDate:      outDate,
			RoomNumber:   int32(numberOfRoom),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(resResp.HotelId) == 0 {
			str = "Failed. Not right reservation information."
		}
	}
	res := map[string]interface{}{
		"message": str,
	}
	json.NewEncoder(w).Encode(res)

}

// return a geoJSON response that allows google map to plot points directly on map
// https://developers.google.com/maps/documentation/javascript/datalayer#sample_geojson
func geoJSONResponse(hs []*profile.Hotel) map[string]interface{} {
	fs := []interface{}{}

	for _, h := range hs {
		fs = append(fs, map[string]interface{}{
			"type": "Feature",
			"id":   h.Id,
			"properties": map[string]interface{}{
				"name":         h.Name,
				"phone_number": h.PhoneNumber,
				"price":        h.Price,
				"score":        h.Score,
				"scoreTimes":   h.ScoreTimes,
			},
			"geometry": map[string]interface{}{
				"type": "Point",
				"coordinates": []float32{
					h.Address.Lon,
					h.Address.Lat,
				},
			},
		})
	}

	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": fs,
	}
}

func checkDataFormat(date string) bool {
	if len(date) != 10 {
		return false
	}
	for i := 0; i < 10; i++ {
		if i == 4 || i == 7 {
			if date[i] != '-' {
				return false
			}
		} else {
			if date[i] < '0' || date[i] > '9' {
				return false
			}
		}
	}
	return true
}
