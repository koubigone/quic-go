package integrationtests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client tests", func() {
	var client *http.Client
	supportedVersions := append([]protocol.VersionNumber{}, protocol.SupportedVersions...)

	BeforeEach(func() {
		err := os.Setenv("HOSTALIASES", "quic.clemente.io 127.0.0.1")
		Expect(err).ToNot(HaveOccurred())
		addr, err := net.ResolveUDPAddr("udp4", "quic.clemente.io:0")
		Expect(err).ToNot(HaveOccurred())
		if addr.String() != "127.0.0.1:0" {
			Fail("quic.clemente.io does not resolve to 127.0.0.1. Consider adding it to /etc/hosts.")
		}
		client = &http.Client{
			Transport: &h2quic.RoundTripper{},
		}
	})

	AfterEach(func() {
		protocol.SupportedVersions = supportedVersions
	})

	for _, v := range supportedVersions {
		version := v

		Context(fmt.Sprintf("with quic version %d", version), func() {
			BeforeEach(func() {
				protocol.SupportedVersions = []protocol.VersionNumber{version}
			})

			It("downloads a hello", func(done Done) {
				resp, err := client.Get("https://quic.clemente.io:" + port + "/hello")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(body)).To(Equal("Hello, World!\n"))
				close(done)
			}, 3)

			It("downloads a small file", func(done Done) {
				dataMan.GenerateData(dataLen)
				resp, err := client.Get("https://quic.clemente.io:" + port + "/data")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(Equal(dataMan.GetData()))
				close(done)
			}, 5)

			It("downloads a large file", func(done Done) {
				dataMan.GenerateData(dataLongLen)
				resp, err := client.Get("https://quic.clemente.io:" + port + "/data")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(Equal(dataMan.GetData()))
				close(done)
			}, 20)

			It("uploads a file", func(done Done) {
				dataMan.GenerateData(dataLen)
				data := bytes.NewReader(dataMan.GetData())
				resp, err := client.Post("https://quic.clemente.io:"+port+"/echo", "text/plain", data)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(Equal(dataMan.GetData()))
				close(done)
			}, 5)
		})
	}
})
