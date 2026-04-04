package dev.kylebradshaw.task.config;

import dev.kylebradshaw.task.security.JwtService;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class JwtConfig {

    @Bean
    public JwtService jwtService(
            @Value("${app.jwt.secret}") String secret,
            @Value("${app.jwt.access-token-ttl-ms:900000}") long accessTtl,
            @Value("${app.jwt.refresh-token-ttl-ms:604800000}") long refreshTtl) {
        return new JwtService(secret, accessTtl, refreshTtl);
    }
}
