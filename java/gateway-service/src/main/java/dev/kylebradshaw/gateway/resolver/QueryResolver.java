package dev.kylebradshaw.gateway.resolver;

import dev.kylebradshaw.gateway.client.ActivityServiceClient;
import dev.kylebradshaw.gateway.client.NotificationServiceClient;
import dev.kylebradshaw.gateway.client.TaskServiceClient;
import dev.kylebradshaw.gateway.dto.ActivityEventDto;
import dev.kylebradshaw.gateway.dto.CommentDto;
import dev.kylebradshaw.gateway.dto.NotificationDto;
import dev.kylebradshaw.gateway.dto.ProjectDto;
import dev.kylebradshaw.gateway.dto.TaskDto;
import graphql.schema.DataFetchingEnvironment;
import org.springframework.graphql.data.method.annotation.Argument;
import org.springframework.graphql.data.method.annotation.QueryMapping;
import org.springframework.stereotype.Controller;

import java.util.List;

@Controller
public class QueryResolver {

    private final TaskServiceClient taskClient;
    private final ActivityServiceClient activityClient;
    private final NotificationServiceClient notificationClient;

    public QueryResolver(TaskServiceClient taskClient,
                         ActivityServiceClient activityClient,
                         NotificationServiceClient notificationClient) {
        this.taskClient = taskClient;
        this.activityClient = activityClient;
        this.notificationClient = notificationClient;
    }

    @QueryMapping
    public List<ProjectDto> myProjects(DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        return taskClient.getMyProjects(userId);
    }

    @QueryMapping
    public ProjectDto project(@Argument String id, DataFetchingEnvironment env) {
        return taskClient.getProject(id);
    }

    @QueryMapping
    public TaskDto task(@Argument String id, DataFetchingEnvironment env) {
        return taskClient.getTask(id);
    }

    @QueryMapping
    public List<ActivityEventDto> taskActivity(@Argument String taskId, DataFetchingEnvironment env) {
        return activityClient.getActivityByTask(taskId);
    }

    @QueryMapping
    public List<CommentDto> taskComments(@Argument String taskId, DataFetchingEnvironment env) {
        return activityClient.getCommentsByTask(taskId);
    }

    @QueryMapping
    public NotificationDto myNotifications(@Argument Boolean unreadOnly, DataFetchingEnvironment env) {
        String userId = env.getGraphQlContext().get("userId");
        return notificationClient.getNotifications(userId, unreadOnly);
    }
}
