package swaggerdocs

import (
	"context"
	"io"
	"net/http"
	"shared/loggers"
	"sync"
	"time"
)

type ServiceSwaggerDocs struct {
	DocsMap sync.Map
	Logger  *loggers.Logger
}

func NewServiceSwaggerDocs(logger *loggers.Logger) *ServiceSwaggerDocs {
	return &ServiceSwaggerDocs{
		Logger: logger,
	}
}

type InfoServices struct {
	Name string
	Url  string
}

func (s *ServiceSwaggerDocs) GetDocs() []InfoServices {
	listInfoServices := []InfoServices{
		{Name: "Auth Service API", Url: "/docs/api?service=auth"},
		{Name: "Budget Planner API", Url: "/docs/api?service=budget"},
	}
	return listInfoServices
}

func (s *ServiceSwaggerDocs) GetDocsAPI(service string) ([]byte, error) {
	if service == "" {
		var docsData []byte
		s.DocsMap.Range(func(key, value any) bool {
			data, ok := value.([]byte)
			if ok {
				docsData = append(docsData, data...)
			} else {
				s.Logger.Error("failed to assertion type docs api: " + key.(string))
			}
			return true
		})
		return docsData, nil
	}
	docsData, found := s.DocsMap.Load(service)
	if !found {
		return nil, ErrNoyFoundDocs
	}
	return docsData.([]byte), nil
}

type ServiceList struct {
	Service string
	Url     string
}

func (s *ServiceSwaggerDocs) UpdateDocs() {
	s.DocsMap.Clear()
	listServices := []ServiceList{
		{Service: "auth", Url: "http://app-auth-user:8080/swager/doc.json"},
		{Service: "budget", Url: "http://app-budget-planner:8080/swager/doc.json"},
	}
	var wg sync.WaitGroup
	for _, list := range listServices {
		wg.Add(1)
		go func(ls ServiceList) {
			defer wg.Done()
			resp, errGetDocs := http.Get(ls.Url)
			if errGetDocs != nil {
				s.Logger.Error("failed to get docs: " + ls.Url + ": " + errGetDocs.Error())
				return
			}
			defer func() {
				if errClose := resp.Body.Close(); errClose != nil {
					s.Logger.Error("failed to close: " + ls.Url + ": " + errClose.Error())
				}
			}()
			dataDocs, errRead := io.ReadAll(resp.Body)
			if errRead != nil {
				s.Logger.Error("failed to read docs: " + ls.Url + ": " + errRead.Error())
				return
			}
			s.DocsMap.Store(list.Service, dataDocs)
		}(list)
	}
	wg.Wait()
}

func (s *ServiceSwaggerDocs) PlanUpdateDocs(ctxCancel context.Context) {
	ticker := time.NewTicker(time.Hour * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ctxCancel.Done():
			s.Logger.Info("graceful shutdown PlanUpdateDocs")
			return
		case <-ticker.C:
			s.UpdateDocs()
			s.Logger.Info("scheduled docs update")
		}
	}
}
