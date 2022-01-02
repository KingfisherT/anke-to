package session

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/srinathgs/mysqlstore"
	"github.com/traPtitech/anke-to/model"
)

type Store struct {
	store *mysqlstore.MySQLStore
}

func (s *Store) GetMiddleware() echo.MiddlewareFunc {
	return session.Middleware(s.store)
}

func (s *Store) GetSession(c echo.Context) (*Session, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return nil, fmt.Errorf("failed to get session:%w", err)
	}

	return &Session{
		c:    c,
		sess: sess,
	}, nil
}

func NewStore(sess model.Session) (*Store, error) {
	store, err := sess.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &Store{store: store}, nil
}

type Session struct {
	c     echo.Context
	sess  *sessions.Session
}

func (s *Session) SetUserID(userID string) {
	s.sess.Values["userID"] = userID
}

func (s *Session) GetUserID() string {
	userID, ok := s.sess.Values["userID"].(string)
	if !ok || userID == "" {
		return ""
	}

	return userID
}

func (s *Session) SetVerifier(verifier string) {
	s.sess.Values["verifier"] = verifier
}

func (s *Session) GetVerifier() string  {
	verifier,ok := s.sess.Values["verifier"].(string)
	if !ok || verifier == "" {
		return ""
	}

	return verifier
}