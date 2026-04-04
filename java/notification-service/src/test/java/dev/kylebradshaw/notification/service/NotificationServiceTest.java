package dev.kylebradshaw.notification.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyDouble;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import dev.kylebradshaw.notification.dto.Notification;
import java.util.LinkedHashSet;
import java.util.Set;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ValueOperations;
import org.springframework.data.redis.core.ZSetOperations;

@ExtendWith(MockitoExtension.class)
class NotificationServiceTest {
    @Mock private StringRedisTemplate redisTemplate;
    @Mock private ZSetOperations<String, String> zSetOps;
    @Mock private ValueOperations<String, String> valueOps;
    @InjectMocks private NotificationService notificationService;

    @Test
    void addNotification_addsToSortedSetAndIncrementsCount() {
        String userId = "user-123";
        when(redisTemplate.opsForZSet()).thenReturn(zSetOps);
        when(redisTemplate.opsForValue()).thenReturn(valueOps);

        var notification = Notification.create("TASK_ASSIGNED", "You were assigned a task", "task-1");
        notificationService.addNotification(userId, notification);

        verify(zSetOps).add(eq("notifications:" + userId), anyString(), anyDouble());
        verify(valueOps).increment("notification_count:" + userId);
    }

    @Test
    void getNotifications_returnsFromRedis() {
        String userId = "user-123";
        when(redisTemplate.opsForZSet()).thenReturn(zSetOps);
        when(redisTemplate.opsForValue()).thenReturn(valueOps);
        when(valueOps.get("notification_count:" + userId)).thenReturn("2");

        String json = "{\"id\":\"1\",\"type\":\"TASK_ASSIGNED\",\"message\":\"Assigned\",\"taskId\":\"t1\",\"read\":false,\"createdAt\":\"2026-04-03T00:00:00Z\"}";
        Set<String> set = new LinkedHashSet<>();
        set.add(json);
        when(zSetOps.reverseRange("notifications:" + userId, 0, -1)).thenReturn(set);

        var response = notificationService.getNotifications(userId, false);
        assertThat(response.unreadCount()).isEqualTo(2);
        assertThat(response.notifications()).hasSize(1);
    }
}
