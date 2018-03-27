package river

import (
	"net/http"

	"github.com/BaritoLog/go-boilerplate/errkit"
	"github.com/BaritoLog/go-boilerplate/httpkit/heartbeat"
	"github.com/gorilla/mux"
)

type receiverUpstream struct {
	addr      string
	appSecret string
	timberCh  chan Timber
	errCh     chan error
}

type ReceiverUpstreamConfig struct {
	Addr      string
	AppSecret string
}

func NewReceiverUpstream(v interface{}) (Upstream, error) {
	conf, ok := v.(ReceiverUpstreamConfig)
	if !ok {
		return nil, errkit.Error("Parameter must be ReceiverUpstreamConfig")
	}

	upstream := &receiverUpstream{
		addr:      conf.Addr,
		appSecret: conf.AppSecret,
		timberCh:  make(chan Timber),
		errCh:     make(chan error),
	}
	return upstream, nil
}

func (u *receiverUpstream) StartTransport() {
	server := &http.Server{
		Addr:    u.addr,
		Handler: u.router(),
	}

	u.errCh <- server.ListenAndServe()
}

func (u *receiverUpstream) TimberChannel() chan Timber {
	return u.timberCh

}
func (u *receiverUpstream) SetErrorChannel(errCh chan error) {
	u.errCh = errCh
}

func (u *receiverUpstream) ErrorChannel() chan error {
	return u.errCh
}

func (u *receiverUpstream) router() (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc(
		"/str/{stream_id}/st/{store_id}/fw/{forwarder_id}/cl/{client_id}/produce/{topic}",
		u.produceHandler,
	).Methods("POST")
	router.HandleFunc("/check-health", heartbeat.Handler)

	return
}

func (u *receiverUpstream) produceHandler(writer http.ResponseWriter, req *http.Request) {

	appSecret := req.Header.Get("Application-Secret")

	if appSecret != u.appSecret {
		http.Error(writer, "Application secret is not valid", http.StatusUnauthorized)
		return
	}

	u.timberCh <- NewTimberFromRequest(req)

	writer.WriteHeader(http.StatusOK)

}