package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"

	"github.com/paw-digital/Pawnimals/server/controller"
	"github.com/paw-digital/Pawnimals/server/image"
	"github.com/paw-digital/Pawnimals/server/net"
	"github.com/paw-digital/Pawnimals/server/spc"
	"github.com/paw-digital/Pawnimals/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	socketio "github.com/googollee/go-socket.io"
	"github.com/jasonlvhit/gocron"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "https://natricon.com"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With, ResponseType")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}

func RandFiles(count int, seed string) {
	if _, err := os.Stat("randsvg"); os.IsNotExist(err) {
		os.Mkdir("randsvg", os.FileMode(0755))
	}
	for i := 0; i < count; i++ {
		address := utils.GenerateAddress()
		sha256 := utils.AddressSha256(address, seed)

		accessories, _ := image.GetAccessoriesForHash(sha256, spc.BTNone, false, nil)
		svg, _ := image.CombineSVG(accessories)
		ioutil.WriteFile(fmt.Sprintf("randsvg/%s.svg", address), svg, os.FileMode(0644))
	}
}

func main() {
	// Get seed from env
	seed := utils.GetEnv("NATRICON_SEED", "0123456789")
	// Parse server options
	loadFiles := flag.Bool("load-files", false, "Print assets as GO arrays")
	testBodyDist := flag.Bool("test-bd", false, "Test body distribution")
	testHairDist := flag.Bool("test-hd", false, "Test hair distribution")
	randomFiles := flag.Int("rand-files", -1, "Generate this many random SVGs and output to randsvg folder")

	serverHost := flag.String("host", "127.0.0.1", "Host to listen on")
	serverPort := flag.Int("port", 8080, "Port to listen on")
	rpcUrl := flag.String("rpc-url", "", "Optional URL to use for nano RPC Client")
	wsUrl := flag.String("nano-ws-url", "", "Nano WS Url to use for tracking donation account")
	flag.Parse()

	if *loadFiles {
		LoadAssetsToArray()
		return
	} else if *randomFiles > 0 {
		fmt.Printf("Generating %d files in ./randsvg", *randomFiles)
		RandFiles(*randomFiles, seed)
		return
	}

	if *testBodyDist {
		controller.TestBodyDistribution(seed)
		return
	} else if *testHairDist {
		controller.TestHairDistribution(seed)
		return
	}

	var rpcClient *net.RPCClient
	if *rpcUrl != "" {
		glog.Infof("RPC Client configured at %s", *rpcUrl)
		rpcClient = &net.RPCClient{Url: *rpcUrl}
	}

	// Setup magickwand
	imagick.Initialize()
	defer imagick.Terminate()

	// Setup router
	router := gin.Default()
	router.Use(CorsMiddleware())

	// Setup socket IO server
	sio := socketio.NewServer(nil)
	sio.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		s.Join("bcast")
		clientId, err := strconv.Atoi(s.ID())
		if err != nil {
			clientId = rand.Intn(1000)
		}
		s.Emit("connected", strconv.Itoa(clientId))
		return nil
	})
	go sio.Serve()
	defer sio.Close()
	router.GET("/socket.io/*any", gin.WrapH(sio))
	router.POST("/socket.io/*any", gin.WrapH(sio))

	// Setup channel for stats processing job
	statsChan := make(chan *gin.Context, 100)

	// Setup natricon controller
	natriconController := controller.NatriconController{
		Seed:         seed,
		StatsChannel: &statsChan,
	}
	// Setup nano controller
	nanoController := controller.NanoController{
		RPCClient:       rpcClient,
		SIOServer:       sio,
		DonationAccount: utils.GetEnv("DONATION_ACCOUNT", ""),
	}

	// V1 API
	router.GET("/api/v1/nano", natriconController.GetNano)
	router.GET("/api/v1/nano/nonce", natriconController.GetNonce)
	// Stats
	router.GET("/api/v1/nano/stats", controller.Stats)
	if gin.IsDebugging() {
		// For testing
		router.GET("/api/natricon", natriconController.GetNatricon)
		router.GET("/api/random", natriconController.GetRandom)
		router.GET("/api/randomsvg", natriconController.GetRandomSvg)
	}

	// Setup cron jobs
	if !gin.IsDebugging() {
		go func() {
			// Checking missed donations
			gocron.Every(30).Minutes().Do(nanoController.CheckMissedCallbacks)
			// Updating principal rep requirement
			gocron.Every(30).Minutes().Do(nanoController.UpdatePrincipalWeight)
			// Update principal reps, this is heavier so dont do it so often
			gocron.Every(30).Minutes().Do(nanoController.UpdatePrincipalReps)
			<-gocron.Start()
		}()
	}

	// Start Nano WS client
	donationAccount := utils.GetEnv("DONATION_ACCOUNT", "")
	if *wsUrl != "" && utils.ValidateAddress(donationAccount) {
		fmt.Printf("\r\nDonation account: %s\r\n", donationAccount)
		go net.StartNanoWSClient(*wsUrl, donationAccount, nanoController.Callback)
	} else {
		fmt.Printf("No donation account specified!\r\n")
	}
	
	// Show wallet
	wallet := utils.GetEnv("WALLET_ID", "")
	if wallet != "" {
		fmt.Printf("Wallet account specified: %s\r\n", wallet)
	} else {
		fmt.Printf("No wallet specified!\r\n")
	}

	// Start stats worker
	go controller.StatsWorker(statsChan)

	// Run on 8080
	router.Run(fmt.Sprintf("%s:%d", *serverHost, *serverPort))
}
