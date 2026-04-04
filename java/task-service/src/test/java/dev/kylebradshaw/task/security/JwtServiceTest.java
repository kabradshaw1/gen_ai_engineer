package dev.kylebradshaw.task.security;

import io.jsonwebtoken.ExpiredJwtException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class JwtServiceTest {

    private static final String SECRET = "test-secret-key-that-is-at-least-32-bytes-long!!";
    private static final long ACCESS_TTL_MS = 900_000L;   // 15 minutes
    private static final long REFRESH_TTL_MS = 604_800_000L; // 7 days

    private JwtService jwtService;

    @BeforeEach
    void setUp() {
        jwtService = new JwtService(SECRET, ACCESS_TTL_MS, REFRESH_TTL_MS);
    }

    @Test
    void generateAndExtractUserId_roundtrip() {
        UUID userId = UUID.randomUUID();
        String email = "user@example.com";

        String token = jwtService.generateAccessToken(userId, email);
        UUID extracted = jwtService.extractUserId(token);

        assertThat(extracted).isEqualTo(userId);
    }

    @Test
    void generateAndExtractEmail_roundtrip() {
        UUID userId = UUID.randomUUID();
        String email = "user@example.com";

        String token = jwtService.generateAccessToken(userId, email);
        String extracted = jwtService.extractEmail(token);

        assertThat(extracted).isEqualTo(email);
    }

    @Test
    void isValid_returnsTrueForValidToken() {
        UUID userId = UUID.randomUUID();
        String token = jwtService.generateAccessToken(userId, "valid@example.com");

        assertThat(jwtService.isValid(token)).isTrue();
    }

    @Test
    void expiredToken_throwsExpiredJwtException() {
        // TTL of 0ms means it expires immediately
        JwtService shortLivedService = new JwtService(SECRET, 0L, REFRESH_TTL_MS);
        UUID userId = UUID.randomUUID();
        String token = shortLivedService.generateAccessToken(userId, "expired@example.com");

        assertThatThrownBy(() -> shortLivedService.extractUserId(token))
                .isInstanceOf(ExpiredJwtException.class);
    }

    @Test
    void generateRefreshTokenString_isValidUUID() {
        String refreshToken = jwtService.generateRefreshTokenString();

        // UUID.fromString throws if not a valid UUID
        UUID parsed = UUID.fromString(refreshToken);
        assertThat(parsed).isNotNull();
    }

    @Test
    void getRefreshTokenTtlMs_returnsConfiguredValue() {
        assertThat(jwtService.getRefreshTokenTtlMs()).isEqualTo(REFRESH_TTL_MS);
    }
}
