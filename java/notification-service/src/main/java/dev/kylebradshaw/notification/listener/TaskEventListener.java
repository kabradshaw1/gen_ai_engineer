package dev.kylebradshaw.notification.listener;

import dev.kylebradshaw.notification.config.RabbitConfig;
import dev.kylebradshaw.notification.dto.Notification;
import dev.kylebradshaw.notification.dto.TaskEventMessage;
import dev.kylebradshaw.notification.service.NotificationService;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Component;

@Component
public class TaskEventListener {

    private final NotificationService notificationService;

    public TaskEventListener(NotificationService notificationService) {
        this.notificationService = notificationService;
    }

    @RabbitListener(queues = RabbitConfig.QUEUE_NAME)
    public void handleTaskEvent(TaskEventMessage message) {
        String taskId = message.taskId() != null ? message.taskId().toString() : null;
        String actorId = message.actorId() != null ? message.actorId().toString() : null;

        switch (message.eventType()) {
            case "TASK_ASSIGNED" -> {
                String assigneeId = extractString(message, "assigneeId");
                if (assigneeId != null) {
                    String taskTitle = extractString(message, "taskTitle");
                    String msg = taskTitle != null
                            ? "You were assigned to task: " + taskTitle
                            : "You were assigned a task";
                    notificationService.addNotification(
                            assigneeId, Notification.create("TASK_ASSIGNED", msg, taskId));
                }
            }
            case "STATUS_CHANGED" -> {
                if (actorId != null) {
                    String taskTitle = extractString(message, "taskTitle");
                    String newStatus = extractString(message, "newStatus");
                    String msg = taskTitle != null && newStatus != null
                            ? "Task \"" + taskTitle + "\" status changed to " + newStatus
                            : "A task status was updated";
                    notificationService.addNotification(
                            actorId, Notification.create("STATUS_CHANGED", msg, taskId));
                }
            }
            case "TASK_CREATED" -> {
                if (actorId != null) {
                    String taskTitle = extractString(message, "taskTitle");
                    String msg = taskTitle != null
                            ? "Task created: " + taskTitle
                            : "A new task was created";
                    notificationService.addNotification(
                            actorId, Notification.create("TASK_CREATED", msg, taskId));
                }
            }
            default -> { /* unhandled event types are intentionally ignored */ }
        }
    }

    private String extractString(TaskEventMessage message, String key) {
        if (message.data() == null) {
            return null;
        }
        Object value = message.data().get(key);
        return value != null ? value.toString() : null;
    }
}
