package dev.kylebradshaw.notification.dto;

import java.time.Instant;
import java.util.UUID;

public record Notification(String id, String type, String message, String taskId, boolean read, Instant createdAt) {
    public static Notification create(String type, String message, String taskId) {
        return new Notification(UUID.randomUUID().toString(), type, message, taskId, false, Instant.now());
    }
}
