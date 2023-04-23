package test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

type TestSuite struct {
	suite.Suite
	ctx context.Context
	ct  compose.ComposeStack
}

func (s *TestSuite) SetupSuite() {
	ct, err := compose.NewDockerComposeWith(
		compose.WithStackFiles("./docker-compose.yml"),
		compose.StackIdentifier("app-test"),
	)
	s.NoError(err, "new compose")
	err = ct.
		Up(s.ctx, compose.Wait(true), compose.RemoveOrphans(true))
	s.NoError(err, "compose up")
	s.ct = ct
}

func (s *TestSuite) TearDownSuite() {
	if s.ct != compose.ComposeStack(nil) {
		err := s.ct.Down(s.ctx,
			compose.RemoveOrphans(true), compose.RemoveVolumes(true))
		s.NoError(err, "compose down")
	}
}

func (s *TestSuite) TestServer() {
	app, err := s.ct.ServiceContainer(s.ctx, "app")
	s.NoError(err)
	ip, err := app.Host(s.ctx)
	s.NoError(err)
	port, err := app.MappedPort(s.ctx, "8080")
	s.NoError(err)
	appAddr := fmt.Sprintf("http://%s:%s", ip, port.Port())

	tests := []struct {
		content [][]byte
	}{
		{
			content: [][]byte{
				[]byte(`{"timestamp":1682257621,"host":"d0.cc","flux":20}`),
				[]byte(`{"timestamp":1682257622,"host":"d0.cc","flux":20}`),
				[]byte(`{"timestamp":1682257623,"host":"d1.cc","flux":20}`),
				[]byte(`{"timestamp":1682257624,"host":"d1.cc","flux":20}`),
			},
		},
		{
			content: [][]byte{
				[]byte(`{"timestamp":1682257621,"host":"d1.cc","flux":20}`),
				[]byte(`{"timestamp":1682257622,"host":"d1.cc","flux":20}`),
				[]byte(`{"timestamp":1682257623,"host":"d0.cc","flux":20}`),
				[]byte(`{"timestamp":1682257624,"host":"d0.cc","flux":20}`),
			},
		},
	}
	send := func(reader io.Reader) (*http.Response, error) {
		return http.Post(appAddr+"/incr", "text/jsonl", reader)
	}
	for _, tt := range tests {
		content := append(bytes.Join(tt.content, []byte("\n")), byte('\n'))
		rd := bytes.NewReader(content)
		resp, err := send(rd)
		s.NoError(err)
		s.Equal(200, resp.StatusCode)
	}

	// check
	u, err := url.Parse(appAddr + "/query")
	s.NoError(err)
	values := u.Query()
	values.Add("time", "202304231345")
	u.RawQuery = values.Encode()
	func() {
		resp, err := http.Get(u.String())
		s.NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(200, resp.StatusCode)
		bs, err := io.ReadAll(resp.Body)
		s.NoError(err)
		s.Equal("d0.cc:80\nd1.cc:80\n", string(bs))
	}()
	values.Add("host", "d0.cc")
	u.RawQuery = values.Encode()
	func() {
		resp, err := http.Get(u.String())
		s.NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(200, resp.StatusCode)
		bs, err := io.ReadAll(resp.Body)
		s.NoError(err)
		s.Equal("d0.cc:80\n", string(bs))
	}()

}

func TestInCompose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	suite.Run(t, &TestSuite{
		ctx: context.Background(),
	})
}
