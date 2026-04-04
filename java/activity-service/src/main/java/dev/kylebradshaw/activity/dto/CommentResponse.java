package dev.kylebradshaw.activity.dto;

import dev.kylebradshaw.activity.document.Comment;
import java.time.Instant;

public record CommentResponse(String id, String taskId, String authorId, String body, Instant createdAt) {
    public static CommentResponse from(Comment comment) {
        return new CommentResponse(comment.getId(), comment.getTaskId(), comment.getAuthorId(), comment.getBody(), comment.getCreatedAt());
    }
}
