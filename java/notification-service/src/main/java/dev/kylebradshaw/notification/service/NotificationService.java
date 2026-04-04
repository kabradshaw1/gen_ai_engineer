package dev.kylebradshaw.notification.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import dev.kylebradshaw.notification.dto.Notification;
import dev.kylebradshaw.notification.dto.NotificationResponse;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

@Service
public class NotificationService {

    private final StringRedisTemplate redisTemplate;
    private final ObjectMapper objectMapper;

    public NotificationService(StringRedisTemplate redisTemplate) {
        this.redisTemplate = redisTemplate;
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.objectMapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
    }

    public void addNotification(String userId, Notification notification) {
        try {
            String json = objectMapper.writeValueAsString(notification);
            double score = notification.createdAt().toEpochMilli();
            redisTemplate.opsForZSet().add("notifications:" + userId, json, score);
            redisTemplate.opsForValue().increment("notification_count:" + userId);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Failed to serialize notification", e);
        }
    }

    public NotificationResponse getNotifications(String userId, boolean unreadOnly) {
        Set<String> entries = redisTemplate.opsForZSet()
                .reverseRange("notifications:" + userId, 0, -1);

        List<Notification> notifications = new ArrayList<>();
        if (entries != null) {
            for (String json : entries) {
                try {
                    Notification n = objectMapper.readValue(json, Notification.class);
                    if (!unreadOnly || !n.read()) {
                        notifications.add(n);
                    }
                } catch (JsonProcessingException e) {
                    // skip malformed entries
                }
            }
        }

        String countStr = redisTemplate.opsForValue().get("notification_count:" + userId);
        long unreadCount = countStr != null ? Long.parseLong(countStr) : 0L;

        return new NotificationResponse(notifications, unreadCount);
    }

    public void markRead(String userId, String notificationId) {
        Set<String> entries = redisTemplate.opsForZSet()
                .reverseRange("notifications:" + userId, 0, -1);

        if (entries == null) {
            return;
        }

        for (String json : entries) {
            try {
                Notification n = objectMapper.readValue(json, Notification.class);
                if (n.id().equals(notificationId) && !n.read()) {
                    redisTemplate.opsForZSet().remove("notifications:" + userId, json);
                    Notification updated = new Notification(
                            n.id(), n.type(), n.message(), n.taskId(), true, n.createdAt());
                    String updatedJson = objectMapper.writeValueAsString(updated);
                    redisTemplate.opsForZSet().add(
                            "notifications:" + userId, updatedJson,
                            n.createdAt().toEpochMilli());
                    redisTemplate.opsForValue().decrement("notification_count:" + userId);
                    return;
                }
            } catch (JsonProcessingException e) {
                // skip malformed entries
            }
        }
    }

    public void markAllRead(String userId) {
        redisTemplate.opsForValue().set("notification_count:" + userId, "0");
    }
}
