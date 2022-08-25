package pushgateway

import (
	"context"
	"net/http"

	"github.com/H-BF/corlib/app/jobs"
	"github.com/H-BF/corlib/pkg/tm"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/push"
)

//NewJob creates new prometheus push-gateway job
func NewJob(ctx context.Context, conf Config) (jobs.JobScheduler, error) {
	const api = "push-gateway.NewPushJob"

	if err := conf.Validate(); err != nil {
		return nil, errors.Wrap(err, api)
	}
	constr := pushGwJobConstructor{Config: &conf}
	ret, err := constr.construct(ctx)
	return ret, errors.Wrap(err, api)
}

//====================================================== IMPL ====================================================

type pushGwJobConstructor struct {
	*Config
}

type wrapHTTPDoer struct {
	ctx     context.Context
	wrapped push.HTTPDoer
}

//Do overrides push.HTTPDoer
func (doer *wrapHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	req = req.WithContext(doer.ctx)
	return doer.wrapped.Do(req)
}

func (maker *pushGwJobConstructor) job(ctx context.Context) error {
	anURL, e := maker.GwEndpointURL(ctx)
	if e != nil {
		return e
	}
	pusher := push.New(anURL, maker.JobName)
	for _, g := range maker.Gatherers {
		_ = pusher.Gatherer(g)
	}
	for _, c := range maker.Collectors {
		_ = pusher.Collector(c)
	}
	if maker.AuthProvider != nil {
		var auth Auth
		if auth, e = maker.AuthProvider(ctx); e != nil {
			return e
		}
		switch a := auth.(type) {
		case NullAuth:
		case UserBasicAuth:
			_ = pusher.BasicAuth(string(a.Username), string(a.Password))
		default:
			return errors.New("unknown 'Auth' is provided")
		}
	}
	if len(maker.ExpFmt) > 0 {
		pusher = pusher.Format(maker.ExpFmt)
	}
	client := maker.HttpClient
	if client == nil {
		client = new(http.Client)
	}
	if maker.RequestDuration > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, maker.RequestDuration)
		defer cancel()
	}
	_ = pusher.Client(&wrapHTTPDoer{ctx: ctx, wrapped: client})

	switch maker.Strategy {
	case PushStrategy:
		e = pusher.Push()
	case AddStrategy:
		e = pusher.Add()
	case DelStrategy:
		e = pusher.Delete()
	default:
		e = errors.Errorf("unknown strategy (%v)", maker.Strategy)
	}
	return e
}

func (maker *pushGwJobConstructor) construct(appCtx context.Context) (jobs.JobScheduler, error) {
	jobConf := jobs.JobSchedulerConf{
		JobID:         maker.JobName,
		TaskScheduler: maker.JobScheduler,
		Backoff:       maker.JobBackoff,
		TaskManager:   maker.TaskManagerProvider,
	}

	jobConf.NewTask = func(ctx context.Context) (tm.Task, []interface{}, error) {
		task, err := tm.MakeSimpleTask(maker.JobName, maker.job)
		return task, []interface{}{ctx}, err
	}

	return jobs.NewJobScheduler(appCtx, jobConf)
}

var _ = NewJob
