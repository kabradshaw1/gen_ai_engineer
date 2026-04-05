# Frontend Updates & Password Auth Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update portfolio site with bios, fix Mermaid diagrams, add global header, implement password auth for Java tasks, and rename repo references.

**Architecture:** Frontend-only changes for tasks 1-4 and 8. Password auth (tasks 5-7) adds backend endpoints to the existing task-service and new frontend components. All auth endpoints return the same JWT format as Google OAuth, so existing token handling is reused.

**Tech Stack:** Next.js 16 + TypeScript, Spring Boot + JPA, BCrypt, Resend Java SDK (`com.resend:resend-java`), Mermaid.js

**Spec:** `docs/superpowers/specs/2026-04-05-frontend-updates-design.md`

---

## File Structure

### Frontend (create/modify)

| Action | File | Responsibility |
|--------|------|---------------|
| Modify | `frontend/src/app/page.tsx` | Home bio text |
| Modify | `frontend/src/app/ai/page.tsx` | AI bio text, tech stack, remove inline nav |
| Modify | `frontend/src/app/java/page.tsx` | Java bio text, tech stack, remove inline nav |
| Modify | `frontend/src/components/MermaidDiagram.tsx` | Fix unique ID generation |
| Move | `frontend/src/components/java/SiteHeader.tsx` to `frontend/src/components/SiteHeader.tsx` | Global header |
| Modify | `frontend/src/app/layout.tsx` | Add AuthProvider + SiteHeader |
| Modify | `frontend/src/app/java/layout.tsx` | Remove SiteHeader + AuthProvider (now in root) |
| Modify | `frontend/src/components/java/AuthProvider.tsx` | Add loginWithPassword and register methods |
| Modify | `frontend/src/components/java/TasksPageContent.tsx` | Add email/password login form |
| Create | `frontend/src/components/java/RegisterForm.tsx` | Registration form |
| Create | `frontend/src/components/java/ForgotPasswordForm.tsx` | Forgot password form |
| Create | `frontend/src/app/java/tasks/reset-password/page.tsx` | Reset password page |

### Java Backend (create/modify)

| Action | File | Responsibility |
|--------|------|---------------|
| Modify | `java/task-service/build.gradle` | Add Resend SDK dependency |
| Modify | `java/task-service/src/main/resources/application.yml` | Add Resend config, frontend URL |
| Modify | `java/task-service/src/main/java/dev/kylebradshaw/task/entity/User.java` | Add passwordHash field |
| Create | `java/task-service/src/main/java/dev/kylebradshaw/task/entity/PasswordResetToken.java` | Reset token entity |
| Create | `java/task-service/src/main/java/dev/kylebradshaw/task/repository/PasswordResetTokenRepository.java` | Reset token repository |
| Create | `java/task-service/src/main/java/dev/kylebradshaw/task/dto/RegisterRequest.java` | Registration DTO |
| Create | `java/task-service/src/main/java/dev/kylebradshaw/task/dto/LoginRequest.java` | Login DTO |
| Create | `java/task-service/src/main/java/dev/kylebradshaw/task/dto/ForgotPasswordRequest.java` | Forgot password DTO |
| Create | `java/task-service/src/main/java/dev/kylebradshaw/task/dto/ResetPasswordRequest.java` | Reset password DTO |
| Create | `java/task-service/src/main/java/dev/kylebradshaw/task/service/EmailService.java` | Resend email sending |
| Create | `java/task-service/src/main/java/dev/kylebradshaw/task/config/ResendConfig.java` | Resend bean |
| Modify | `java/task-service/src/main/java/dev/kylebradshaw/task/service/AuthService.java` | Add register, login, forgot/reset password |
| Modify | `java/task-service/src/main/java/dev/kylebradshaw/task/controller/AuthController.java` | Add new endpoints |
| Modify | `java/task-service/src/main/java/dev/kylebradshaw/task/config/SecurityConfig.java` | Add PasswordEncoder bean |

### Repo Rename (modify)

| Action | File | What changes |
|--------|------|-------------|
| Modify | `docker-compose.yml` | GHCR image paths (3 refs) |
| Modify | `k8s/ai-services/deployments/ingestion.yml` | GHCR image path |
| Modify | `k8s/ai-services/deployments/chat.yml` | GHCR image path |
| Modify | `k8s/ai-services/deployments/debug.yml` | GHCR image path |
| Modify | `java/k8s/deployments/task-service.yml` | GHCR image path |
| Modify | `java/k8s/deployments/gateway-service.yml` | GHCR image path |
| Modify | `java/k8s/deployments/activity-service.yml` | GHCR image path |
| Modify | `java/k8s/deployments/notification-service.yml` | GHCR image path |
| Modify | `Makefile` | SSH path |
| Modify | `k8s/setup-windows.ps1` | Default repo path |
| Modify | `CLAUDE.md` | Path references |

### Tests (create/modify)

| Action | File | Responsibility |
|--------|------|---------------|
| Create | `java/task-service/src/test/java/dev/kylebradshaw/task/service/EmailServiceTest.java` | EmailService tests |
| Modify | `java/task-service/src/test/java/dev/kylebradshaw/task/service/AuthServiceTest.java` | Add password auth tests |
| Modify | `java/task-service/src/test/java/dev/kylebradshaw/task/controller/AuthControllerTest.java` | Add endpoint tests |

---

## Task 1: Fix Mermaid Diagrams

**Files:**
- Modify: `frontend/src/components/MermaidDiagram.tsx`

- [ ] **Step 1: Fix the MermaidDiagram component to use unique IDs**

Replace the entire file content:

```tsx
"use client";

import { useEffect, useId, useRef } from "react";
import mermaid from "mermaid";
import DOMPurify from "dompurify";

mermaid.initialize({
  startOnLoad: false,
  theme: "dark",
  themeVariables: {
    darkMode: true,
  },
});

let counter = 0;

export function MermaidDiagram({ chart }: { chart: string }) {
  const ref = useRef<HTMLDivElement>(null);
  const reactId = useId();

  useEffect(() => {
    if (!ref.current) return;

    // useId() returns strings like ":r0:" which contain colons that are
    // invalid in mermaid element IDs. Combine with a counter for uniqueness.
    const id = `mermaid-${reactId.replace(/:/g, "")}-${counter++}`;
    mermaid.render(id, chart).then(({ svg }) => {
      if (ref.current) {
        ref.current.innerHTML = DOMPurify.sanitize(svg);  // sanitized via DOMPurify
      }
    });
  }, [chart, reactId]);

  return <div ref={ref} className="flex justify-center overflow-x-auto" />;
}
```

- [ ] **Step 2: Verify the fix locally**

Run: `cd frontend && npx tsc --noEmit`
Expected: No type errors.

Then open `http://localhost:3000/ai` in the browser and verify both diagrams render with visible text.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/MermaidDiagram.tsx
git commit -m "fix: use unique IDs for Mermaid diagrams to fix multi-diagram rendering"
```

---

## Task 2: Update Bios and Tech Stack

**Files:**
- Modify: `frontend/src/app/page.tsx`
- Modify: `frontend/src/app/ai/page.tsx`
- Modify: `frontend/src/app/java/page.tsx`

- [ ] **Step 1: Update home page bio**

In `frontend/src/app/page.tsx`, replace the placeholder paragraph (lines 17-19):

```tsx
// Old:
          [Placeholder: General bio. Introduce yourself, your background in
          software engineering, and what you bring to the table. This page serves
          as the entry point to role-specific sections below.]

// New:
          Software engineer focused on building production systems with modern
          tooling. Since August 2022, I&apos;ve been working full-time on
          personal projects and consulting, with a focus on Go, TypeScript, and
          cloud-native infrastructure. This portfolio showcases two areas of
          specialization — AI/ML engineering and full-stack Java development.
```

- [ ] **Step 2: Update AI page bio and tech stack**

In `frontend/src/app/ai/page.tsx`, replace the placeholder paragraph (lines 59-61):

```tsx
// Old:
            [Placeholder: AI-focused bio. Describe your interest in generative
            AI, RAG architectures, and building intelligent systems. Highlight
            relevant experience and what excites you about the field.]

// New:
            Building intelligent systems with retrieval-augmented generation and
            agentic architectures. This section demonstrates RAG pipelines,
            vector search, LLM orchestration, and tool-using agents — built with
            FastAPI, Qdrant, and Ollama, deployed on Kubernetes.
```

Update the Document Q&A tech stack (line 82):

```tsx
// Old:
            <li>Docker Compose orchestration</li>

// New:
            <li>Minikube Kubernetes deployment (production), Docker Compose (local dev)</li>
```

Add a bullet to the Debug Assistant tech stack, after the SSE streaming line (after line 122):

```tsx
            <li>Minikube Kubernetes deployment (production)</li>
```

- [ ] **Step 3: Update Java page bio and tech stack**

In `frontend/src/app/java/page.tsx`, replace the placeholder paragraph (lines 17-19):

```tsx
// Old:
          [Placeholder: Java-focused bio. Describe your experience with Spring
          Boot, microservices, and cloud-native development. Highlight relevant
          projects and what excites you about backend engineering at scale.]

// New:
          Full-stack microservices architecture with Spring Boot, GraphQL, and
          event-driven communication. This section demonstrates a task
          management platform built with four Java services, PostgreSQL,
          MongoDB, Redis, and RabbitMQ — deployed on Kubernetes.
```

Update the tech stack (line 40):

```tsx
// Old:
          <li>Docker Compose + Minikube Kubernetes manifests</li>

// New:
          <li>Minikube Kubernetes deployment (production), Docker Compose (local dev)</li>
```

- [ ] **Step 4: Verify**

Run: `cd frontend && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/page.tsx frontend/src/app/ai/page.tsx frontend/src/app/java/page.tsx
git commit -m "feat: add bios and update tech stack to reflect Minikube deployment"
```

---

## Task 3: Global Header

**Files:**
- Move: `frontend/src/components/java/SiteHeader.tsx` to `frontend/src/components/SiteHeader.tsx`
- Modify: `frontend/src/app/layout.tsx`
- Modify: `frontend/src/app/java/layout.tsx`
- Modify: `frontend/src/app/ai/page.tsx`
- Modify: `frontend/src/app/java/page.tsx`

- [ ] **Step 1: Move and update SiteHeader**

Create `frontend/src/components/SiteHeader.tsx` with the updated content (delete the old file after):

```tsx
"use client";

import Link from "next/link";
import { FileText } from "lucide-react";
import { useAuth } from "@/components/java/AuthProvider";
import { NotificationBell } from "@/components/java/NotificationBell";

export function SiteHeader() {
  const { user, isLoggedIn, logout } = useAuth();

  return (
    <header className="border-b border-foreground/10 bg-background">
      <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-6">
        <Link href="/" className="text-lg font-semibold">
          Kyle Bradshaw
        </Link>

        <nav className="flex items-center gap-4">
          <a
            href="https://github.com/kabradshaw1/portfolio"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            Portfolio
          </a>
          <a
            href="https://github.com/kabradshaw1"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            GitHub
          </a>
          <a
            href="https://www.linkedin.com/in/kyle-bradshaw-15950988/"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            LinkedIn
          </a>
          <a
            href="/resume.pdf"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            <FileText className="size-5" />
          </a>

          {isLoggedIn && (
            <>
              <NotificationBell />
              <div className="flex items-center gap-2">
                {user?.avatarUrl && (
                  <img
                    src={user.avatarUrl}
                    alt=""
                    className="size-7 rounded-full"
                  />
                )}
                <span className="text-sm text-muted-foreground">
                  {user?.name}
                </span>
                <button
                  onClick={logout}
                  className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  Sign out
                </button>
              </div>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
```

Delete the old file:
```bash
rm frontend/src/components/java/SiteHeader.tsx
```

- [ ] **Step 2: Update root layout to include AuthProvider and SiteHeader**

Replace `frontend/src/app/layout.tsx`:

```tsx
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/components/java/AuthProvider";
import { SiteHeader } from "@/components/SiteHeader";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Kyle Bradshaw",
  description: "Portfolio and project showcase",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased dark`}
    >
      <body className="min-h-full flex flex-col">
        <AuthProvider>
          <SiteHeader />
          {children}
        </AuthProvider>
      </body>
    </html>
  );
}
```

- [ ] **Step 3: Simplify Java layout**

Replace `frontend/src/app/java/layout.tsx`:

```tsx
export default function JavaLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}
```

- [ ] **Step 4: Remove inline nav links from AI and Java pages**

In `frontend/src/app/ai/page.tsx`, remove the navigation Link block (lines 46-51):

```tsx
// Remove this block:
        <Link
          href="/"
          className="text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          &larr; Home
        </Link>
```

In `frontend/src/app/java/page.tsx`, remove the navigation Link block (lines 6-11):

```tsx
// Remove this block:
      <Link
        href="/"
        className="text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        &larr; Home
      </Link>
```

- [ ] **Step 5: Verify and clean up imports**

Search for any file importing from `@/components/java/SiteHeader` and update to `@/components/SiteHeader`. The Java layout was the only consumer, and we already updated it.

Run: `cd frontend && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/SiteHeader.tsx frontend/src/app/layout.tsx frontend/src/app/java/layout.tsx frontend/src/app/ai/page.tsx frontend/src/app/java/page.tsx
git rm frontend/src/components/java/SiteHeader.tsx
git commit -m "feat: move SiteHeader to global layout, add portfolio repo link"
```

---

## Task 4: Password Auth Backend — DTOs, Entity, Repository

**Files:**
- Modify: `java/task-service/build.gradle`
- Modify: `java/task-service/src/main/resources/application.yml`
- Modify: `java/task-service/src/main/java/dev/kylebradshaw/task/entity/User.java`
- Create: `java/task-service/src/main/java/dev/kylebradshaw/task/entity/PasswordResetToken.java`
- Create: `java/task-service/src/main/java/dev/kylebradshaw/task/repository/PasswordResetTokenRepository.java`
- Create: `java/task-service/src/main/java/dev/kylebradshaw/task/dto/RegisterRequest.java`
- Create: `java/task-service/src/main/java/dev/kylebradshaw/task/dto/LoginRequest.java`
- Create: `java/task-service/src/main/java/dev/kylebradshaw/task/dto/ForgotPasswordRequest.java`
- Create: `java/task-service/src/main/java/dev/kylebradshaw/task/dto/ResetPasswordRequest.java`
- Modify: `java/task-service/src/main/java/dev/kylebradshaw/task/config/SecurityConfig.java`

- [ ] **Step 1: Add Resend SDK dependency**

In `java/task-service/build.gradle`, add after the jjwt-jackson line (line 10):

```gradle
    implementation 'com.resend:resend-java:+'
```

- [ ] **Step 2: Add Resend and frontend URL config**

In `java/task-service/src/main/resources/application.yml`, add under the `app:` section (after line 31):

```yaml
  resend:
    api-key: ${RESEND_API_KEY:}
    from-email: ${RESEND_FROM_EMAIL:onboarding@resend.dev}
  frontend-url: ${FRONTEND_URL:http://localhost:3000}
```

- [ ] **Step 3: Add passwordHash field to User entity**

In `java/task-service/src/main/java/dev/kylebradshaw/task/entity/User.java`, add after the `avatarUrl` field (after line 26):

```java
    @Column(name = "password_hash")
    private String passwordHash;
```

Add a second constructor after the existing one (after line 37):

```java
    public User(String email, String name, String passwordHash, boolean isPasswordUser) {
        this.email = email;
        this.name = name;
        this.passwordHash = passwordHash;
    }
```

Add getter and setter before `getCreatedAt()` (before line 63):

```java
    public String getPasswordHash() {
        return passwordHash;
    }

    public void setPasswordHash(String passwordHash) {
        this.passwordHash = passwordHash;
    }
```

- [ ] **Step 4: Create PasswordResetToken entity**

Create `java/task-service/src/main/java/dev/kylebradshaw/task/entity/PasswordResetToken.java`:

```java
package dev.kylebradshaw.task.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "password_reset_tokens")
public class PasswordResetToken {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(nullable = false, unique = true)
    private String token;

    @ManyToOne
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @Column(name = "expires_at", nullable = false)
    private Instant expiresAt;

    protected PasswordResetToken() {}

    public PasswordResetToken(String token, User user, Instant expiresAt) {
        this.token = token;
        this.user = user;
        this.expiresAt = expiresAt;
    }

    public UUID getId() {
        return id;
    }

    public String getToken() {
        return token;
    }

    public User getUser() {
        return user;
    }

    public Instant getExpiresAt() {
        return expiresAt;
    }

    public boolean isExpired() {
        return Instant.now().isAfter(expiresAt);
    }
}
```

- [ ] **Step 5: Create PasswordResetTokenRepository**

Create `java/task-service/src/main/java/dev/kylebradshaw/task/repository/PasswordResetTokenRepository.java`:

```java
package dev.kylebradshaw.task.repository;

import dev.kylebradshaw.task.entity.PasswordResetToken;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;
import java.util.UUID;

public interface PasswordResetTokenRepository extends JpaRepository<PasswordResetToken, UUID> {
    Optional<PasswordResetToken> findByToken(String token);

    void deleteByUserId(UUID userId);
}
```

- [ ] **Step 6: Create DTO records**

Create `java/task-service/src/main/java/dev/kylebradshaw/task/dto/RegisterRequest.java`:

```java
package dev.kylebradshaw.task.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;

public record RegisterRequest(
        @NotBlank @Email String email,
        @NotBlank @Size(min = 8) String password,
        @NotBlank String name
) {}
```

Create `java/task-service/src/main/java/dev/kylebradshaw/task/dto/LoginRequest.java`:

```java
package dev.kylebradshaw.task.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

public record LoginRequest(
        @NotBlank @Email String email,
        @NotBlank String password
) {}
```

Create `java/task-service/src/main/java/dev/kylebradshaw/task/dto/ForgotPasswordRequest.java`:

```java
package dev.kylebradshaw.task.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

public record ForgotPasswordRequest(
        @NotBlank @Email String email
) {}
```

Create `java/task-service/src/main/java/dev/kylebradshaw/task/dto/ResetPasswordRequest.java`:

```java
package dev.kylebradshaw.task.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;

public record ResetPasswordRequest(
        @NotBlank String token,
        @NotBlank @Size(min = 8) String password
) {}
```

- [ ] **Step 7: Add PasswordEncoder bean to SecurityConfig**

In `java/task-service/src/main/java/dev/kylebradshaw/task/config/SecurityConfig.java`, add the imports:

```java
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
```

Add the bean method after the `corsConfigurationSource()` method (before the closing `}`):

```java
    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }
```

- [ ] **Step 8: Verify compilation**

Run: `cd java && ./gradlew :task-service:compileJava`
Expected: BUILD SUCCESSFUL

- [ ] **Step 9: Commit**

```bash
git add java/task-service/build.gradle \
  java/task-service/src/main/resources/application.yml \
  java/task-service/src/main/java/dev/kylebradshaw/task/entity/User.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/entity/PasswordResetToken.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/repository/PasswordResetTokenRepository.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/dto/RegisterRequest.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/dto/LoginRequest.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/dto/ForgotPasswordRequest.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/dto/ResetPasswordRequest.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/config/SecurityConfig.java
git commit -m "feat: add password auth DTOs, entities, and PasswordEncoder bean"
```

---

## Task 5: Password Auth Backend — EmailService

**Files:**
- Create: `java/task-service/src/main/java/dev/kylebradshaw/task/service/EmailService.java`
- Create: `java/task-service/src/main/java/dev/kylebradshaw/task/config/ResendConfig.java`
- Create: `java/task-service/src/test/java/dev/kylebradshaw/task/service/EmailServiceTest.java`

- [ ] **Step 1: Write the EmailService test**

Create `java/task-service/src/test/java/dev/kylebradshaw/task/service/EmailServiceTest.java`:

```java
package dev.kylebradshaw.task.service;

import com.resend.Resend;
import com.resend.core.exception.ResendException;
import com.resend.services.emails.Emails;
import com.resend.services.emails.model.CreateEmailOptions;
import com.resend.services.emails.model.CreateEmailResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class EmailServiceTest {

    @Mock
    private Resend resend;

    @Mock
    private Emails emails;

    private EmailService emailService;

    @BeforeEach
    void setUp() {
        when(resend.emails()).thenReturn(emails);
        emailService = new EmailService(resend, "noreply@resend.dev", "https://kylebradshaw.dev");
    }

    @Test
    void sendPasswordResetEmail_sendsWithCorrectParams() throws ResendException {
        CreateEmailResponse mockResponse = new CreateEmailResponse();
        when(emails.send(any(CreateEmailOptions.class))).thenReturn(mockResponse);

        emailService.sendPasswordResetEmail("user@example.com", "reset-token-123");

        ArgumentCaptor<CreateEmailOptions> captor = ArgumentCaptor.forClass(CreateEmailOptions.class);
        verify(emails).send(captor.capture());

        CreateEmailOptions options = captor.getValue();
        assertThat(options.getFrom()).isEqualTo("noreply@resend.dev");
        assertThat(options.getSubject()).isEqualTo("Reset your password");
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd java && ./gradlew :task-service:test --tests "dev.kylebradshaw.task.service.EmailServiceTest" --no-daemon`
Expected: FAIL — `EmailService` class does not exist.

- [ ] **Step 3: Implement EmailService**

Create `java/task-service/src/main/java/dev/kylebradshaw/task/service/EmailService.java`:

```java
package dev.kylebradshaw.task.service;

import com.resend.Resend;
import com.resend.core.exception.ResendException;
import com.resend.services.emails.model.CreateEmailOptions;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

@Service
public class EmailService {

    private static final Logger log = LoggerFactory.getLogger(EmailService.class);

    private final Resend resend;
    private final String fromEmail;
    private final String frontendUrl;

    public EmailService(
            Resend resend,
            @Value("${app.resend.from-email:onboarding@resend.dev}") String fromEmail,
            @Value("${app.frontend-url:http://localhost:3000}") String frontendUrl) {
        this.resend = resend;
        this.fromEmail = fromEmail;
        this.frontendUrl = frontendUrl;
    }

    public void sendPasswordResetEmail(String toEmail, String token) {
        String resetUrl = frontendUrl + "/java/tasks/reset-password?token=" + token;
        String subject = "Reset your password";
        String body = "<h2>Reset your password</h2>"
                + "<p>Click the link below to reset your password. This link expires in 1 hour.</p>"
                + "<p><a href=\"" + resetUrl + "\">Reset Password</a></p>"
                + "<p>If you didn't request this, you can ignore this email.</p>";

        CreateEmailOptions options = CreateEmailOptions.builder()
                .from(fromEmail)
                .to(toEmail)
                .subject(subject)
                .html(body)
                .build();

        try {
            resend.emails().send(options);
        } catch (ResendException e) {
            log.error("Failed to send password reset email to {}", toEmail, e);
            throw new RuntimeException("Failed to send password reset email", e);
        }
    }
}
```

- [ ] **Step 4: Create Resend bean configuration**

Create `java/task-service/src/main/java/dev/kylebradshaw/task/config/ResendConfig.java`:

```java
package dev.kylebradshaw.task.config;

import com.resend.Resend;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class ResendConfig {

    @Bean
    public Resend resend(@Value("${app.resend.api-key:}") String apiKey) {
        return new Resend(apiKey);
    }
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd java && ./gradlew :task-service:test --tests "dev.kylebradshaw.task.service.EmailServiceTest" --no-daemon`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add java/task-service/src/main/java/dev/kylebradshaw/task/service/EmailService.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/config/ResendConfig.java \
  java/task-service/src/test/java/dev/kylebradshaw/task/service/EmailServiceTest.java
git commit -m "feat: add EmailService with Resend SDK for password reset emails"
```

---

## Task 6: Password Auth Backend — AuthService + Controller

**Files:**
- Modify: `java/task-service/src/main/java/dev/kylebradshaw/task/service/AuthService.java`
- Modify: `java/task-service/src/main/java/dev/kylebradshaw/task/controller/AuthController.java`
- Modify: `java/task-service/src/test/java/dev/kylebradshaw/task/service/AuthServiceTest.java`
- Modify: `java/task-service/src/test/java/dev/kylebradshaw/task/controller/AuthControllerTest.java`

- [ ] **Step 1: Write AuthService tests for password auth**

In `java/task-service/src/test/java/dev/kylebradshaw/task/service/AuthServiceTest.java`, add imports at the top:

```java
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import dev.kylebradshaw.task.entity.PasswordResetToken;
import dev.kylebradshaw.task.repository.PasswordResetTokenRepository;
```

Add new mock fields after the existing `@Mock` fields (after line 38):

```java
    @Mock
    private PasswordResetTokenRepository passwordResetTokenRepository;

    @Mock
    private EmailService emailService;

    private final PasswordEncoder passwordEncoder = new BCryptPasswordEncoder();
```

Update the `setUp()` method (line 44):

```java
    @BeforeEach
    void setUp() {
        authService = new AuthService(userRepository, refreshTokenRepository,
                jwtService, passwordEncoder, passwordResetTokenRepository, emailService);
    }
```

Add the following tests after the existing tests (after line 145):

```java
    @Test
    void register_newUser_createsUserAndReturnsTokens() {
        String email = "new@example.com";
        String name = "New User";
        String password = "password123";

        when(userRepository.findByEmail(email)).thenReturn(Optional.empty());
        when(userRepository.save(any(User.class))).thenAnswer(inv -> inv.getArgument(0));
        when(jwtService.generateAccessToken(any(), eq(email))).thenReturn("access-token");
        when(jwtService.generateRefreshTokenString()).thenReturn("refresh-token");
        when(jwtService.getRefreshTokenTtlMs()).thenReturn(604_800_000L);
        when(refreshTokenRepository.save(any())).thenAnswer(inv -> inv.getArgument(0));

        AuthResponse response = authService.register(email, password, name);

        assertThat(response.accessToken()).isEqualTo("access-token");
        assertThat(response.email()).isEqualTo(email);
        assertThat(response.name()).isEqualTo(name);
        verify(userRepository).save(argThat(u ->
                u.getEmail().equals(email) && u.getPasswordHash() != null));
    }

    @Test
    void register_existingEmail_throwsException() {
        String email = "existing@example.com";
        when(userRepository.findByEmail(email)).thenReturn(Optional.of(new User(email, "User", null)));

        assertThatThrownBy(() -> authService.register(email, "password123", "User"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("already registered");
    }

    @Test
    void login_validCredentials_returnsTokens() {
        String email = "user@example.com";
        String password = "password123";
        String hashedPassword = passwordEncoder.encode(password);

        User user = new User(email, "User", hashedPassword, true);
        when(userRepository.findByEmail(email)).thenReturn(Optional.of(user));
        when(jwtService.generateAccessToken(any(), eq(email))).thenReturn("access-token");
        when(jwtService.generateRefreshTokenString()).thenReturn("refresh-token");
        when(jwtService.getRefreshTokenTtlMs()).thenReturn(604_800_000L);
        when(refreshTokenRepository.save(any())).thenAnswer(inv -> inv.getArgument(0));

        AuthResponse response = authService.login(email, password);

        assertThat(response.accessToken()).isEqualTo("access-token");
        assertThat(response.email()).isEqualTo(email);
    }

    @Test
    void login_wrongPassword_throwsException() {
        String email = "user@example.com";
        User user = new User(email, "User", passwordEncoder.encode("correct"), true);
        when(userRepository.findByEmail(email)).thenReturn(Optional.of(user));

        assertThatThrownBy(() -> authService.login(email, "wrong"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Invalid");
    }

    @Test
    void login_noAccount_throwsException() {
        when(userRepository.findByEmail("nobody@example.com")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> authService.login("nobody@example.com", "password"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Invalid");
    }

    @Test
    void forgotPassword_existingUser_sendsEmail() {
        String email = "user@example.com";
        User user = new User(email, "User", "hash", true);
        when(userRepository.findByEmail(email)).thenReturn(Optional.of(user));
        when(passwordResetTokenRepository.save(any())).thenAnswer(inv -> inv.getArgument(0));

        authService.forgotPassword(email);

        verify(emailService).sendPasswordResetEmail(eq(email), any(String.class));
    }

    @Test
    void forgotPassword_nonExistentUser_doesNotThrow() {
        when(userRepository.findByEmail("nobody@example.com")).thenReturn(Optional.empty());

        // Should not throw — don't reveal whether email exists
        authService.forgotPassword("nobody@example.com");

        verify(emailService, never()).sendPasswordResetEmail(any(), any());
    }

    @Test
    void resetPassword_validToken_updatesPassword() {
        String tokenStr = UUID.randomUUID().toString();
        User user = new User("user@example.com", "User", "old-hash", true);
        PasswordResetToken resetToken = new PasswordResetToken(
                tokenStr, user, Instant.now().plusSeconds(3600));

        when(passwordResetTokenRepository.findByToken(tokenStr)).thenReturn(Optional.of(resetToken));

        authService.resetPassword(tokenStr, "newpassword123");

        assertThat(passwordEncoder.matches("newpassword123", user.getPasswordHash())).isTrue();
        verify(passwordResetTokenRepository).delete(resetToken);
    }

    @Test
    void resetPassword_expiredToken_throwsException() {
        String tokenStr = UUID.randomUUID().toString();
        User user = new User("user@example.com", "User", "hash", true);
        PasswordResetToken expiredToken = new PasswordResetToken(
                tokenStr, user, Instant.now().minusSeconds(1));

        when(passwordResetTokenRepository.findByToken(tokenStr)).thenReturn(Optional.of(expiredToken));

        assertThatThrownBy(() -> authService.resetPassword(tokenStr, "newpassword"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("expired");
    }

    @Test
    void resetPassword_invalidToken_throwsException() {
        when(passwordResetTokenRepository.findByToken("bad-token")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> authService.resetPassword("bad-token", "newpassword"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Invalid");
    }
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd java && ./gradlew :task-service:test --tests "dev.kylebradshaw.task.service.AuthServiceTest" --no-daemon`
Expected: FAIL — `register`, `login`, `forgotPassword`, `resetPassword` methods don't exist.

- [ ] **Step 3: Implement AuthService password methods**

Replace `java/task-service/src/main/java/dev/kylebradshaw/task/service/AuthService.java` entirely:

```java
package dev.kylebradshaw.task.service;

import dev.kylebradshaw.task.dto.AuthResponse;
import dev.kylebradshaw.task.entity.PasswordResetToken;
import dev.kylebradshaw.task.entity.RefreshToken;
import dev.kylebradshaw.task.entity.User;
import dev.kylebradshaw.task.repository.PasswordResetTokenRepository;
import dev.kylebradshaw.task.repository.RefreshTokenRepository;
import dev.kylebradshaw.task.repository.UserRepository;
import dev.kylebradshaw.task.security.JwtService;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.UUID;

@Service
public class AuthService {

    private final UserRepository userRepository;
    private final RefreshTokenRepository refreshTokenRepository;
    private final JwtService jwtService;
    private final PasswordEncoder passwordEncoder;
    private final PasswordResetTokenRepository passwordResetTokenRepository;
    private final EmailService emailService;

    public AuthService(UserRepository userRepository,
                       RefreshTokenRepository refreshTokenRepository,
                       JwtService jwtService,
                       PasswordEncoder passwordEncoder,
                       PasswordResetTokenRepository passwordResetTokenRepository,
                       EmailService emailService) {
        this.userRepository = userRepository;
        this.refreshTokenRepository = refreshTokenRepository;
        this.jwtService = jwtService;
        this.passwordEncoder = passwordEncoder;
        this.passwordResetTokenRepository = passwordResetTokenRepository;
        this.emailService = emailService;
    }

    @Transactional
    public AuthResponse authenticateGoogleUser(String email, String name, String avatarUrl) {
        User user = userRepository.findByEmail(email).orElse(null);

        if (user == null) {
            user = new User(email, name, avatarUrl);
            user = userRepository.save(user);
        } else {
            user.setName(name);
            user.setAvatarUrl(avatarUrl);
        }

        return issueTokens(user);
    }

    @Transactional
    public AuthResponse register(String email, String password, String name) {
        if (userRepository.findByEmail(email).isPresent()) {
            throw new IllegalArgumentException("Email already registered");
        }

        String hashedPassword = passwordEncoder.encode(password);
        User user = new User(email, name, hashedPassword, true);
        user = userRepository.save(user);

        return issueTokens(user);
    }

    @Transactional
    public AuthResponse login(String email, String password) {
        User user = userRepository.findByEmail(email)
                .orElseThrow(() -> new IllegalArgumentException("Invalid email or password"));

        if (user.getPasswordHash() == null || !passwordEncoder.matches(password, user.getPasswordHash())) {
            throw new IllegalArgumentException("Invalid email or password");
        }

        return issueTokens(user);
    }

    @Transactional
    public void forgotPassword(String email) {
        userRepository.findByEmail(email).ifPresent(user -> {
            String token = UUID.randomUUID().toString();
            Instant expiresAt = Instant.now().plusSeconds(3600); // 1 hour
            PasswordResetToken resetToken = new PasswordResetToken(token, user, expiresAt);
            passwordResetTokenRepository.save(resetToken);
            emailService.sendPasswordResetEmail(email, token);
        });
    }

    @Transactional
    public void resetPassword(String token, String newPassword) {
        PasswordResetToken resetToken = passwordResetTokenRepository.findByToken(token)
                .orElseThrow(() -> new IllegalArgumentException("Invalid reset token"));

        if (resetToken.isExpired()) {
            throw new IllegalArgumentException("Reset token has expired");
        }

        User user = resetToken.getUser();
        user.setPasswordHash(passwordEncoder.encode(newPassword));
        userRepository.save(user);
        passwordResetTokenRepository.delete(resetToken);
    }

    @Transactional
    public AuthResponse refreshAccessToken(String refreshTokenStr) {
        RefreshToken refreshToken = refreshTokenRepository.findByToken(refreshTokenStr)
                .orElseThrow(() -> new IllegalArgumentException("Refresh token not found"));

        if (refreshToken.isExpired()) {
            throw new IllegalArgumentException("Refresh token is expired");
        }

        User user = refreshToken.getUser();
        return issueTokens(user);
    }

    private AuthResponse issueTokens(User user) {
        String accessToken = jwtService.generateAccessToken(user.getId(), user.getEmail());
        String refreshTokenStr = jwtService.generateRefreshTokenString();
        Instant expiresAt = Instant.now().plusMillis(jwtService.getRefreshTokenTtlMs());

        RefreshToken refreshToken = new RefreshToken(user, refreshTokenStr, expiresAt);
        refreshTokenRepository.save(refreshToken);

        return new AuthResponse(
                accessToken,
                refreshTokenStr,
                user.getId(),
                user.getEmail(),
                user.getName(),
                user.getAvatarUrl()
        );
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd java && ./gradlew :task-service:test --tests "dev.kylebradshaw.task.service.AuthServiceTest" --no-daemon`
Expected: PASS (all existing + new tests)

- [ ] **Step 5: Write AuthController tests for new endpoints**

In `java/task-service/src/test/java/dev/kylebradshaw/task/controller/AuthControllerTest.java`, add imports:

```java
import dev.kylebradshaw.task.dto.RegisterRequest;
import dev.kylebradshaw.task.dto.LoginRequest;
import dev.kylebradshaw.task.dto.ForgotPasswordRequest;
import dev.kylebradshaw.task.dto.ResetPasswordRequest;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.doNothing;
```

Add tests after the existing tests (after line 97):

```java
    @Test
    void register_validRequest_returnsTokens() throws Exception {
        UUID userId = UUID.randomUUID();
        AuthResponse authResponse = new AuthResponse(
                "access-token", "refresh-token", userId, "new@example.com", "New User", null);

        when(authService.register("new@example.com", "password123", "New User"))
                .thenReturn(authResponse);

        String body = objectMapper.writeValueAsString(
                new RegisterRequest("new@example.com", "password123", "New User"));

        mockMvc.perform(post("/api/auth/register")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.accessToken").value("access-token"))
                .andExpect(jsonPath("$.email").value("new@example.com"));
    }

    @Test
    void login_validCredentials_returnsTokens() throws Exception {
        UUID userId = UUID.randomUUID();
        AuthResponse authResponse = new AuthResponse(
                "access-token", "refresh-token", userId, "user@example.com", "User", null);

        when(authService.login("user@example.com", "password123"))
                .thenReturn(authResponse);

        String body = objectMapper.writeValueAsString(
                new LoginRequest("user@example.com", "password123"));

        mockMvc.perform(post("/api/auth/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.accessToken").value("access-token"));
    }

    @Test
    void forgotPassword_returnsNoContent() throws Exception {
        doNothing().when(authService).forgotPassword("user@example.com");

        String body = objectMapper.writeValueAsString(
                new ForgotPasswordRequest("user@example.com"));

        mockMvc.perform(post("/api/auth/forgot-password")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isNoContent());
    }

    @Test
    void resetPassword_returnsNoContent() throws Exception {
        doNothing().when(authService).resetPassword("valid-token", "newpassword123");

        String body = objectMapper.writeValueAsString(
                new ResetPasswordRequest("valid-token", "newpassword123"));

        mockMvc.perform(post("/api/auth/reset-password")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isNoContent());
    }
```

- [ ] **Step 6: Run controller tests to verify they fail**

Run: `cd java && ./gradlew :task-service:test --tests "dev.kylebradshaw.task.controller.AuthControllerTest" --no-daemon`
Expected: FAIL — new endpoints don't exist on the controller.

- [ ] **Step 7: Add new endpoints to AuthController**

In `java/task-service/src/main/java/dev/kylebradshaw/task/controller/AuthController.java`, add imports:

```java
import dev.kylebradshaw.task.dto.RegisterRequest;
import dev.kylebradshaw.task.dto.LoginRequest;
import dev.kylebradshaw.task.dto.ForgotPasswordRequest;
import dev.kylebradshaw.task.dto.ResetPasswordRequest;
import org.springframework.http.ResponseEntity;
```

Add the following endpoints after the existing `refresh` method (after line 90):

```java
    @PostMapping("/register")
    public AuthResponse register(@Valid @RequestBody RegisterRequest request) {
        return authService.register(request.email(), request.password(), request.name());
    }

    @PostMapping("/login")
    public AuthResponse login(@Valid @RequestBody LoginRequest request) {
        return authService.login(request.email(), request.password());
    }

    @PostMapping("/forgot-password")
    public ResponseEntity<Void> forgotPassword(@Valid @RequestBody ForgotPasswordRequest request) {
        authService.forgotPassword(request.email());
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/reset-password")
    public ResponseEntity<Void> resetPassword(@Valid @RequestBody ResetPasswordRequest request) {
        authService.resetPassword(request.token(), request.password());
        return ResponseEntity.noContent().build();
    }
```

- [ ] **Step 8: Run all auth tests**

Run: `cd java && ./gradlew :task-service:test --tests "dev.kylebradshaw.task.controller.AuthControllerTest" --tests "dev.kylebradshaw.task.service.AuthServiceTest" --no-daemon`
Expected: PASS

- [ ] **Step 9: Run full Java preflight**

Run: `make preflight-java`
Expected: All checks pass (checkstyle + unit tests).

- [ ] **Step 10: Commit**

```bash
git add java/task-service/src/main/java/dev/kylebradshaw/task/service/AuthService.java \
  java/task-service/src/main/java/dev/kylebradshaw/task/controller/AuthController.java \
  java/task-service/src/test/java/dev/kylebradshaw/task/service/AuthServiceTest.java \
  java/task-service/src/test/java/dev/kylebradshaw/task/controller/AuthControllerTest.java
git commit -m "feat: add register, login, forgot-password, reset-password endpoints"
```

---

## Task 7: Password Auth Frontend — Login, Register, Forgot Password

**Files:**
- Modify: `frontend/src/components/java/AuthProvider.tsx`
- Modify: `frontend/src/components/java/TasksPageContent.tsx`
- Create: `frontend/src/components/java/RegisterForm.tsx`
- Create: `frontend/src/components/java/ForgotPasswordForm.tsx`
- Create: `frontend/src/app/java/tasks/reset-password/page.tsx`

- [ ] **Step 1: Update AuthProvider with password login and register methods**

Replace `frontend/src/components/java/AuthProvider.tsx`:

```tsx
"use client";

import {
  createContext,
  useCallback,
  useContext,
  useState,
} from "react";
import { ApolloProvider } from "@apollo/client/react";
import { apolloClient } from "@/lib/apollo-client";
import {
  clearTokens,
  isLoggedIn as checkIsLoggedIn,
  setTokens,
  GATEWAY_URL,
} from "@/lib/auth";

interface AuthUser {
  userId: string;
  email: string;
  name: string;
  avatarUrl: string | null;
}

interface AuthContextType {
  user: AuthUser | null;
  isLoggedIn: boolean;
  login: (code: string, redirectUri: string) => Promise<void>;
  loginWithPassword: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  isLoggedIn: false,
  login: async () => {},
  loginWithPassword: async () => {},
  register: async () => {},
  logout: () => {},
});

export function useAuth() {
  return useContext(AuthContext);
}

function handleAuthResponse(data: {
  accessToken: string;
  refreshToken: string;
  userId: string;
  email: string;
  name: string;
  avatarUrl: string | null;
}): AuthUser {
  setTokens(data.accessToken, data.refreshToken);
  const authUser: AuthUser = {
    userId: data.userId,
    email: data.email,
    name: data.name,
    avatarUrl: data.avatarUrl,
  };
  localStorage.setItem("java_user", JSON.stringify(authUser));
  return authUser;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(() => {
    if (typeof window === "undefined" || !checkIsLoggedIn()) return null;
    const stored = localStorage.getItem("java_user");
    return stored ? JSON.parse(stored) : null;
  });
  const [isAuthenticated, setIsAuthenticated] = useState(
    () => typeof window !== "undefined" && checkIsLoggedIn(),
  );

  const login = useCallback(async (code: string, redirectUri: string) => {
    const res = await fetch(`${GATEWAY_URL}/api/auth/google`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code, redirectUri }),
    });
    if (!res.ok) throw new Error("Login failed");
    const data = await res.json();
    const authUser = handleAuthResponse(data);
    setUser(authUser);
    setIsAuthenticated(true);
  }, []);

  const loginWithPassword = useCallback(async (email: string, password: string) => {
    const res = await fetch(`${GATEWAY_URL}/api/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) {
      const errorText = await res.text();
      throw new Error(errorText || "Invalid email or password");
    }
    const data = await res.json();
    const authUser = handleAuthResponse(data);
    setUser(authUser);
    setIsAuthenticated(true);
  }, []);

  const register = useCallback(async (email: string, password: string, name: string) => {
    const res = await fetch(`${GATEWAY_URL}/api/auth/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password, name }),
    });
    if (!res.ok) {
      const errorText = await res.text();
      throw new Error(errorText || "Registration failed");
    }
    const data = await res.json();
    const authUser = handleAuthResponse(data);
    setUser(authUser);
    setIsAuthenticated(true);
  }, []);

  const logout = useCallback(() => {
    clearTokens();
    localStorage.removeItem("java_user");
    setUser(null);
    setIsAuthenticated(false);
    apolloClient.clearStore();
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, isLoggedIn: isAuthenticated, login, loginWithPassword, register, logout }}
    >
      <ApolloProvider client={apolloClient}>{children}</ApolloProvider>
    </AuthContext.Provider>
  );
}
```

- [ ] **Step 2: Update TasksPageContent with email/password login**

Replace `frontend/src/components/java/TasksPageContent.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/components/java/AuthProvider";
import { GoogleLoginButton } from "@/components/java/GoogleLoginButton";
import { RegisterForm } from "@/components/java/RegisterForm";
import { ForgotPasswordForm } from "@/components/java/ForgotPasswordForm";
import { ProjectList } from "@/components/java/ProjectList";
import { useSearchParams } from "next/navigation";

type AuthView = "login" | "register" | "forgot-password";

export function TasksPageContent() {
  const { isLoggedIn, login, loginWithPassword } = useAuth();
  const searchParams = useSearchParams();
  const [view, setView] = useState<AuthView>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const resetSuccess = searchParams.get("reset") === "success";

  // Handle OAuth callback
  useEffect(() => {
    const code = searchParams.get("code");
    if (code && !isLoggedIn) {
      const redirectUri = `${window.location.origin}/java/tasks`;
      login(code, redirectUri).then(() => {
        window.history.replaceState({}, "", "/java/tasks");
      });
    }
  }, [searchParams, isLoggedIn, login]);

  if (isLoggedIn) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-12">
        <ProjectList />
      </div>
    );
  }

  if (view === "register") {
    return (
      <div className="mx-auto max-w-sm px-6 py-24">
        <RegisterForm onBack={() => setView("login")} />
      </div>
    );
  }

  if (view === "forgot-password") {
    return (
      <div className="mx-auto max-w-sm px-6 py-24">
        <ForgotPasswordForm onBack={() => setView("login")} />
      </div>
    );
  }

  const handlePasswordLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await loginWithPassword(email, password);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-sm px-6 py-24">
      <div className="flex flex-col gap-6">
        <div className="text-center">
          <h1 className="text-2xl font-semibold">Task Manager</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Sign in to manage your projects and tasks.
          </p>
        </div>

        {resetSuccess && (
          <p className="text-sm text-green-500 text-center">
            Password reset successful. You can now sign in.
          </p>
        )}

        {error && (
          <p className="text-sm text-red-500 text-center">{error}</p>
        )}

        <form onSubmit={handlePasswordLogin} className="flex flex-col gap-4">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
          />
          <button
            type="submit"
            disabled={loading}
            className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            {loading ? "Signing in..." : "Sign in"}
          </button>
        </form>

        <div className="flex justify-between text-sm">
          <button
            onClick={() => setView("forgot-password")}
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Forgot password?
          </button>
          <button
            onClick={() => setView("register")}
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Create account
          </button>
        </div>

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-foreground/10" />
          </div>
          <div className="relative flex justify-center text-xs">
            <span className="bg-background px-2 text-muted-foreground">or</span>
          </div>
        </div>

        <div className="flex justify-center">
          <GoogleLoginButton />
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Create RegisterForm component**

Create `frontend/src/components/java/RegisterForm.tsx`:

```tsx
"use client";

import { useState } from "react";
import { useAuth } from "@/components/java/AuthProvider";

export function RegisterForm({ onBack }: { onBack: () => void }) {
  const { register } = useAuth();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }

    setLoading(true);
    try {
      await register(email, password, name);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="text-center">
        <h1 className="text-2xl font-semibold">Create Account</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Sign up to start managing tasks.
        </p>
      </div>

      {error && (
        <p className="text-sm text-red-500 text-center">{error}</p>
      )}

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <input
          type="text"
          placeholder="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <input
          type="password"
          placeholder="Password (min 8 characters)"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          minLength={8}
          className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <input
          type="password"
          placeholder="Confirm password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          required
          className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <button
          type="submit"
          disabled={loading}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
        >
          {loading ? "Creating account..." : "Create account"}
        </button>
      </form>

      <button
        onClick={onBack}
        className="text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        Already have an account? Sign in
      </button>
    </div>
  );
}
```

- [ ] **Step 4: Create ForgotPasswordForm component**

Create `frontend/src/components/java/ForgotPasswordForm.tsx`:

```tsx
"use client";

import { useState } from "react";
import { GATEWAY_URL } from "@/lib/auth";

export function ForgotPasswordForm({ onBack }: { onBack: () => void }) {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await fetch(`${GATEWAY_URL}/api/auth/forgot-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      if (!res.ok) throw new Error("Request failed");
      setSent(true);
    } catch {
      setError("Something went wrong. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  if (sent) {
    return (
      <div className="flex flex-col gap-6 text-center">
        <h1 className="text-2xl font-semibold">Check your email</h1>
        <p className="text-sm text-muted-foreground">
          If an account exists for {email}, we sent a password reset link.
        </p>
        <button
          onClick={onBack}
          className="text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          Back to sign in
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="text-center">
        <h1 className="text-2xl font-semibold">Forgot password</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Enter your email and we&apos;ll send you a reset link.
        </p>
      </div>

      {error && (
        <p className="text-sm text-red-500 text-center">{error}</p>
      )}

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <button
          type="submit"
          disabled={loading}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
        >
          {loading ? "Sending..." : "Send reset link"}
        </button>
      </form>

      <button
        onClick={onBack}
        className="text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        Back to sign in
      </button>
    </div>
  );
}
```

- [ ] **Step 5: Create reset password page**

Create `frontend/src/app/java/tasks/reset-password/page.tsx`:

```tsx
"use client";

import { useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { GATEWAY_URL } from "@/lib/auth";

export default function ResetPasswordPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const token = searchParams.get("token");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  if (!token) {
    return (
      <div className="mx-auto max-w-sm px-6 py-24 text-center">
        <h1 className="text-2xl font-semibold">Invalid link</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          This password reset link is invalid or has expired.
        </p>
      </div>
    );
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(`${GATEWAY_URL}/api/auth/reset-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, password }),
      });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "Reset failed");
      }
      router.push("/java/tasks?reset=success");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Reset failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-sm px-6 py-24">
      <div className="flex flex-col gap-6">
        <div className="text-center">
          <h1 className="text-2xl font-semibold">Reset password</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Enter your new password.
          </p>
        </div>

        {error && (
          <p className="text-sm text-red-500 text-center">{error}</p>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <input
            type="password"
            placeholder="New password (min 8 characters)"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
          />
          <input
            type="password"
            placeholder="Confirm new password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            className="rounded-md border border-foreground/20 bg-background px-3 py-2 text-sm"
          />
          <button
            type="submit"
            disabled={loading}
            className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            {loading ? "Resetting..." : "Reset password"}
          </button>
        </form>
      </div>
    </div>
  );
}
```

- [ ] **Step 6: Verify frontend compiles**

Run: `cd frontend && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/java/AuthProvider.tsx \
  frontend/src/components/java/TasksPageContent.tsx \
  frontend/src/components/java/RegisterForm.tsx \
  frontend/src/components/java/ForgotPasswordForm.tsx \
  frontend/src/app/java/tasks/reset-password/page.tsx
git commit -m "feat: add password login, registration, and forgot password UI"
```

---

## Task 8: Repo Rename — Update References

**Files:**
- Modify: `docker-compose.yml` (3 GHCR refs)
- Modify: `k8s/ai-services/deployments/ingestion.yml`
- Modify: `k8s/ai-services/deployments/chat.yml`
- Modify: `k8s/ai-services/deployments/debug.yml`
- Modify: `java/k8s/deployments/task-service.yml`
- Modify: `java/k8s/deployments/gateway-service.yml`
- Modify: `java/k8s/deployments/activity-service.yml`
- Modify: `java/k8s/deployments/notification-service.yml`
- Modify: `Makefile`
- Modify: `k8s/setup-windows.ps1`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update GHCR image references in docker-compose.yml**

In `docker-compose.yml`, replace all 3 occurrences:
- `ghcr.io/kabradshaw1/gen_ai_engineer/ingestion` to `ghcr.io/kabradshaw1/portfolio/ingestion`
- `ghcr.io/kabradshaw1/gen_ai_engineer/chat` to `ghcr.io/kabradshaw1/portfolio/chat`
- `ghcr.io/kabradshaw1/gen_ai_engineer/debug` to `ghcr.io/kabradshaw1/portfolio/debug`

- [ ] **Step 2: Update GHCR image references in K8s AI deployments**

In each file, replace `ghcr.io/kabradshaw1/gen_ai_engineer/` with `ghcr.io/kabradshaw1/portfolio/`:

- `k8s/ai-services/deployments/ingestion.yml`
- `k8s/ai-services/deployments/chat.yml`
- `k8s/ai-services/deployments/debug.yml`

- [ ] **Step 3: Update GHCR image references in K8s Java deployments**

In each file, replace `ghcr.io/kabradshaw1/gen_ai_engineer/` with `ghcr.io/kabradshaw1/portfolio/`:

- `java/k8s/deployments/task-service.yml`
- `java/k8s/deployments/gateway-service.yml`
- `java/k8s/deployments/activity-service.yml`
- `java/k8s/deployments/notification-service.yml`

- [ ] **Step 4: Update Makefile SSH path**

In `Makefile`, replace:
- `C:\Users\PC\repos\gen_ai_engineer` to `C:\Users\PC\repos\portfolio`

- [ ] **Step 5: Update setup-windows.ps1**

In `k8s/setup-windows.ps1`, replace the default repo path suggestion:
- `gen_ai_engineer` to `portfolio`

- [ ] **Step 6: Update CLAUDE.md references**

In `CLAUDE.md`, find-and-replace:
- `gen_ai_engineer` to `portfolio` (for all path references)

Review each replacement to ensure it makes sense in context.

- [ ] **Step 7: Verify no remaining references in active files**

Run a grep excluding docs/superpowers/ and node_modules to confirm no stale references remain in active config files.

- [ ] **Step 8: Commit**

```bash
git add docker-compose.yml \
  k8s/ai-services/deployments/ingestion.yml \
  k8s/ai-services/deployments/chat.yml \
  k8s/ai-services/deployments/debug.yml \
  java/k8s/deployments/task-service.yml \
  java/k8s/deployments/gateway-service.yml \
  java/k8s/deployments/activity-service.yml \
  java/k8s/deployments/notification-service.yml \
  Makefile \
  k8s/setup-windows.ps1 \
  CLAUDE.md
git commit -m "chore: rename repo references from gen_ai_engineer to portfolio"
```

---

## Task 9: Run Full Preflight

- [ ] **Step 1: Run full preflight**

Run: `make preflight`
Expected: All checks pass (Python + frontend + security + Java).

- [ ] **Step 2: Fix any failures**

Address any lint, type, or test failures that arise from the changes.
