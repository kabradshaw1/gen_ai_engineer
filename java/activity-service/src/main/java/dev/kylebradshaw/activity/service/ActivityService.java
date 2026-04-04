package dev.kylebradshaw.activity.service;

import dev.kylebradshaw.activity.document.ActivityEvent;
import dev.kylebradshaw.activity.repository.ActivityEventRepository;
import java.util.List;
import org.springframework.stereotype.Service;

@Service
public class ActivityService {
    private final ActivityEventRepository activityRepo;

    public ActivityService(ActivityEventRepository activityRepo) {
        this.activityRepo = activityRepo;
    }

    public List<ActivityEvent> getActivityByTask(String taskId) {
        return activityRepo.findByTaskIdOrderByTimestampDesc(taskId);
    }

    public List<ActivityEvent> getActivityByProject(String projectId) {
        return activityRepo.findByProjectIdOrderByTimestampDesc(projectId);
    }
}
