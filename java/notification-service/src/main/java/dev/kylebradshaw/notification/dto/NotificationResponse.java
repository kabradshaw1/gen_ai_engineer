package dev.kylebradshaw.notification.dto;

import java.util.List;

public record NotificationResponse(List<Notification> notifications, long unreadCount) {}
