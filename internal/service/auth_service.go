package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"registro-despesas/internal/models"
	"registro-despesas/internal/repository"
)

const (
	tokenTTL       = 2 * time.Hour
	minSenha       = 6
	maxSenhaBytes  = 72 // limite do bcrypt
	dummyPlaintext = "senha-inexistente"
)

var (
	ErrCredenciais        = errors.New("e-mail ou senha inválidos")
	ErrSenhaAtual         = errors.New("senha atual incorreta")
	ErrSenhaCurta         = errors.New("a senha deve ter no mínimo 6 caracteres")
	ErrSenhaLonga         = errors.New("a senha deve ter no máximo 72 bytes")
	ErrUsuarioInexistente = errors.New("usuário não encontrado")
)

type AuthService struct {
	repo   *repository.AuthRepository
	secret []byte
	// dummyHash equaliza o tempo de resposta quando o e-mail não existe.
	dummyHash []byte
}

func NewAuthService(repo *repository.AuthRepository, secret string) (*AuthService, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte(dummyPlaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash dummy: %w", err)
	}
	return &AuthService{repo: repo, secret: []byte(secret), dummyHash: dummy}, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	u, err := s.repo.FindByEmail(ctx, strings.TrimSpace(req.Email))
	if errors.Is(err, repository.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(req.Senha))
		return nil, ErrCredenciais
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Senha), []byte(req.Senha)); err != nil {
		return nil, ErrCredenciais
	}

	token, err := s.generateToken(u)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{Token: token, Nome: u.Nome, Email: u.Email, Nivel: u.Nivel}, nil
}

func (s *AuthService) AlterarSenha(ctx context.Context, userID int, req models.AlterarSenhaRequest) error {
	if err := validarSenha(req.SenhaNova); err != nil {
		return err
	}

	u, err := s.repo.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrUsuarioInexistente
	}
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Senha), []byte(req.SenhaAtual)); err != nil {
		return ErrSenhaAtual
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.SenhaNova), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gerar hash: %w", err)
	}

	return s.repo.UpdateSenha(ctx, u.ID, string(hash))
}

// validarSenha aplica as regras de tamanho (mín. 6 caracteres, máx. 72 bytes do bcrypt).
func validarSenha(senha string) error {
	if utf8.RuneCountInString(senha) < minSenha {
		return ErrSenhaCurta
	}
	if len(senha) > maxSenhaBytes {
		return ErrSenhaLonga
	}
	return nil
}

func (s *AuthService) generateToken(u *models.Usuario) (string, error) {
	now := time.Now()
	claims := models.JWTClaims{
		UserID: u.ID,
		Email:  u.Email,
		Nivel:  u.Nivel,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
