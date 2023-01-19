package server

type Option func(*Server)

func EnableAuth(enabled bool) Option {
	return func(s *Server) {
		s.authEnabled = enabled
	}
}

func Prefix(prefix string) Option {
	return func(s *Server) {
		s.prefix = prefix
	}
}
