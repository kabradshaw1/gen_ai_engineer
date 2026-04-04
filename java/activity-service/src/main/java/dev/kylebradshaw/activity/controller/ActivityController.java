package dev.kylebradshaw.activity.controller;

import dev.kylebradshaw.activity.document.ActivityEvent;
import dev.kylebradshaw.activity.service.ActivityService;
import java.util.List;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/activity")
public class ActivityController {
    private final ActivityService activityService;

    public ActivityController(ActivityService activityService) { this.activityService = activityService; }

    @GetMapping("/task/{taskId}")
    public List<ActivityEvent> getByTask(@PathVariable String taskId) {
        return activityService.getActivityByTask(taskId);
    }

    @GetMapping("/project/{projectId}")
    public List<ActivityEvent> getByProject(@PathVariable String projectId) {
        return activityService.getActivityByProject(projectId);
    }
}
