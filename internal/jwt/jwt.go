package jwt

import (
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"math/rand"
	"net/http"
	"time"
)

var (
	JWTSecret       = "your_secret_key"
	TokenExpiration = time.Hour * 1
	CookieName      = "token"
	UserIDKey       = "user"
)

func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Пытаемся получить куку
			cookie, err := c.Cookie(CookieName)

			if err != nil || cookie.Value == "" {
				// Если куки нет или она пустая, создаем новую
				return createAndSetNewToken(c, next)
			}

			// Проверяем существующий токен
			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
				}
				return []byte(JWTSecret), nil
			})

			if err != nil || !token.Valid {
				// Если токен невалидный, создаем новый
				return createAndSetNewToken(c, next)
			}

			// Получаем claims из токена
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// Устанавливаем user_id в контекст для использования в handlers
				c.Set(UserIDKey, claims[UserIDKey])
				return next(c)
			}

			// Если что-то пошло не так, создаем новый токен
			return createAndSetNewToken(c, next)
		}
	}
}

// createAndSetNewToken создает новый JWT-токен и устанавливает его в куку
func createAndSetNewToken(c echo.Context, next echo.HandlerFunc) error {
	// Генерируем новый UUID для пользователя
	userID := genRandomID(6)

	// Создаем новый токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		UserIDKey: userID,
		"exp":     time.Now().Add(TokenExpiration).Unix(),
		"iat":     time.Now().Unix(),
	})

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create token")
	}

	// Создаем новую куку
	cookie := new(http.Cookie)
	cookie.Name = CookieName
	cookie.Value = tokenString
	cookie.Expires = time.Now().Add(TokenExpiration)
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteStrictMode

	// Устанавливаем куку
	c.SetCookie(cookie)

	// Устанавливаем user_id в контекст
	c.Set(UserIDKey, userID)

	return next(c)
}

// Helper finc

func genRandomID(num int) string {
	var letters = []rune("0123456789")

	str := make([]rune, num)
	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}
	return string(str)
}
