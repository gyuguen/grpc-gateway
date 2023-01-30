package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	pb "github.com/grpc-gateway/pb/echo/v1"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const TimeFormat string = "2006-01-02 15:04:05"

type Result struct {
	ty    string
	reqTm time.Time
	rcvTm time.Time
	idx   int
	err   error
}

func TestLimiter(t *testing.T) {
	errCh := make(chan error)

	err := Serve(9090, 8080, errCh)

	require.NoError(t, err)

	log.Info("Wait 1 sec up to server")
	time.Sleep(time.Second * 1)

	log.Info("success run grpc and rest")

	resultChannels := sendGrpc(5)
	//resultChannels := sendRest(1)
	resultChannels = append(resultChannels, sendRest(5)...)

	var results []Result

	for _, ch := range resultChannels {
		results = append(results, <-ch)
	}

	resultMap := make(map[string]int)

	for _, result := range results {
		reqTmStr := result.reqTm.Format(TimeFormat)
		rcvTmStr := result.rcvTm.Format(TimeFormat)
		cnt, ok := resultMap[rcvTmStr]
		if !ok {
			cnt = 0
		}

		cnt++

		resultMap[rcvTmStr] = cnt

		log.Infof("[%s] reqTm(%s) rcvTm(%s), idx(%d), err(%v)",
			result.ty,
			reqTmStr,
			rcvTmStr,
			result.idx,
			result.err,
		)
	}

	for tm, cnt := range resultMap {
		log.Infof("tm(%s), cnt(%d)", tm, cnt)
	}

}

func TestEcho(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := grpc.Dial(
		"20.24.34.167:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	require.NoError(t, err)

	cli := pb.NewEcoServiceClient(conn)

	req := &pb.EchoRequest{
		Name:   "123123123",
		Bearer: "456789",
	}
	resp, err := cli.Echo(ctx, req)
	require.NoError(t, err)
	fmt.Println(resp)
}

func sendGrpc(count int) []chan Result {

	conn, err := grpc.Dial(
		"localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil
	}

	cli := pb.NewEcoServiceClient(conn)

	results := make([]chan Result, count)

	for i := 0; i < count; i++ {
		c := make(chan Result)
		go func(i int) {
			req := &pb.EchoRequest{
				Name: "test",
			}
			log.Infof("[Grpc] %d request exceed", i)
			reqTm := time.Now()
			_, err := cli.Echo(context.Background(), req)
			log.Infof("[Grpc] %d received", i)
			c <- Result{
				ty:    "grpc",
				idx:   i,
				reqTm: reqTm,
				rcvTm: time.Now(),
				err:   err,
			}

		}(i)
		results[i] = c
	}

	return results
}

func sendRest(count int) []chan Result {

	results := make([]chan Result, count)

	for i := 0; i < count; i++ {
		c := make(chan Result)
		go func(i int) {
			body := `{"name": "test"}`
			reqBody := bytes.NewBufferString(body)
			log.Infof("[Rest] %d request exceed", i)
			reqTm := time.Now()
			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v1/echo", reqBody)
			if err != nil {
				c <- Result{
					ty:    "rest",
					idx:   i,
					reqTm: reqTm,
					rcvTm: time.Now(),
					err:   err,
				}
				return
			}
			req.Header["Authorization"] = []string{"tests"}
			cli := http.Client{}
			resp, err := cli.Do(req)
			if err != nil {
				c <- Result{
					ty:    "rest",
					idx:   i,
					reqTm: reqTm,
					rcvTm: time.Now(),
					err:   err,
				}
				return
			}
			res, err := io.ReadAll(resp.Body)
			if err != nil {
				c <- Result{
					ty:    "rest",
					idx:   i,
					reqTm: reqTm,
					rcvTm: time.Now(),
					err:   err,
				}
				return
			}
			log.Infof("Res: %s", string(res))
			//_, err := http.Post("http://localhost:8080", "application/json", reqBody)
			log.Infof("[Rest] %d received", i)
			c <- Result{
				ty:    "rest",
				idx:   i,
				reqTm: reqTm,
				rcvTm: time.Now(),
				err:   err,
			}
		}(i)
		results[i] = c
	}

	return results
}
