package signature_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/go-slack-event-router/internal/testutils"
	"github.com/genkami/go-slack-event-router/signature"
)

var _ = Describe("Signature", func() {
	Describe("Middleware", func() {
		var (
			token        = "THE_TOKEN"
			content      = []byte(`{"body": "this is a request body"}`)
			innerHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			middleware *signature.Middleware
		)

		BeforeEach(func() {
			middleware = &signature.Middleware{
				SigningSecret:   token,
				VerboseResponse: true,
				Handler:         innerHandler,
			}
		})

		Context("when the signature is valid", func() {
			It("calls the inner handler", func() {
				req, err := http.NewRequest(http.MethodPost, "http://example.com/", bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())
				err = testutils.AddSignature(req.Header, []byte(token), content, time.Now())
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when the request is not signed", func() {
			It("responds with BadRequest", func() {
				req, err := http.NewRequest(http.MethodPost, "http://example.com/", bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when the signature is completely wrong", func() {
			It("responds with BadRequest", func() {
				req, err := http.NewRequest(http.MethodPost, "http://example.com/", bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())
				err = testutils.AddSignature(req.Header, []byte(token), content, time.Now())
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set(testutils.HeaderSignature, "WRONG_HEADER")
				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when the request may be signed with wrong token", func() {
			It("responds with Unauthorized", func() {
				req, err := http.NewRequest(http.MethodPost, "http://example.com/", bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())
				err = testutils.AddSignature(req.Header, []byte("OOPS_I_MISTOOK_THE_TOKEN"), content, time.Now())
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when timestamp header is not given", func() {
			It("responds with BadRequest", func() {
				req, err := http.NewRequest(http.MethodPost, "http://example.com/", bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())
				err = testutils.AddSignature(req.Header, []byte(token), content, time.Now())
				Expect(err).NotTo(HaveOccurred())
				req.Header.Del(testutils.HeaderTimestamp)
				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when the timestamp is too old", func() {
			It("responds with BadRequest", func() {
				req, err := http.NewRequest(http.MethodPost, "http://example.com/", bytes.NewReader(content))
				Expect(err).NotTo(HaveOccurred())
				err = testutils.AddSignature(req.Header, []byte(token), content, time.Now().Add(-1*time.Hour))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
