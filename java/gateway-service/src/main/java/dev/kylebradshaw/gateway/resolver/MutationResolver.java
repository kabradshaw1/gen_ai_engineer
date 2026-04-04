package dev.kylebradshaw.gateway.resolver;

import dev.kylebradshaw.gateway.client.ActivityServiceClient;
import dev.kylebradshaw.gateway.client.NotificationServiceClient;
import dev.kylebradshaw.gateway.client.TaskServiceClient;
import dev.kylebradshaw.gateway.dto.CommentDto;
import dev.kylebradshaw.gateway.dto.ProjectDto;
import dev.kylebradshaw.gateway.dto.TaskDto;
import graphql.schema.DataFetchingEnvironment;
import org.springframework.graphql.data.method.annotation.Argument;
import org.springframework.graphql.data.method.annotation.MutationMapping;
import org.springframework.stereotype.Controller;

import java.util.Map;

@Controller
public class MutationResolver {

    private final TaskServiceClient taskClient;
    private final ActivityServiceClient activityClient;
    private final NotificationServiceClient notificationClient;

    public MutationResolver(TaskServiceClient taskClient,
                            ActivityServiceClient activityClient,
                            NotificationServiceClient notificationClient) {
        this.taskClient = taskClient;
        this.activityClient = activityClient;
        this.notificationClient = notificationClient;
    }

    @MutationMapping
    public ProjectDto createProject(@Argument Map<String, Object> input, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        return taskClient.createProject(userId, input);
    }

    @MutationMapping
    public ProjectDto updateProject(@Argument String id, @Argument Map<String, Object> input, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        return taskClient.updateProject(id, userId, input);
    }

    @MutationMapping
    public Boolean deleteProject(@Argument String id, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        taskClient.deleteProject(id, userId);
        return true;
    }

    @MutationMapping
    public TaskDto createTask(@Argument Map<String, Object> input, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        return taskClient.createTask(userId, input);
    }

    @MutationMapping
    public TaskDto updateTask(@Argument String id, @Argument Map<String, Object> input, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        return taskClient.updateTask(id, userId, input);
    }

    @MutationMapping
    public Boolean deleteTask(@Argument String id, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        taskClient.deleteTask(id, userId);
        return true;
    }

    @MutationMapping
    public TaskDto assignTask(@Argument String taskId, @Argument String userId, DataFetchingEnvironment env) {
        String requestingUserId = env.getGraphQlContext().get("userId");
        return taskClient.assignTask(taskId, userId, requestingUserId);
    }

    @MutationMapping
    public CommentDto addComment(@Argument String taskId, @Argument String body, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        return activityClient.addComment(taskId, userId, body);
    }

    @MutationMapping
    public Boolean markNotificationRead(@Argument String id, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        notificationClient.markRead(userId, id);
        return true;
    }

    @MutationMapping
    public Boolean markAllNotificationsRead(DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        notificationClient.markAllRead(userId);
        return true;
    }
}
