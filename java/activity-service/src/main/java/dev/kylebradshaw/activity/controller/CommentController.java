package dev.kylebradshaw.activity.controller;

import dev.kylebradshaw.activity.dto.CommentResponse;
import dev.kylebradshaw.activity.dto.CreateCommentRequest;
import dev.kylebradshaw.activity.service.CommentService;
import jakarta.validation.Valid;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/comments")
public class CommentController {
    private final CommentService commentService;

    public CommentController(CommentService commentService) { this.commentService = commentService; }

    @PostMapping("/{taskId}")
    @ResponseStatus(HttpStatus.CREATED)
    public CommentResponse addComment(@PathVariable String taskId, @RequestHeader("X-User-Id") String userId,
                                       @Valid @RequestBody CreateCommentRequest request) {
        return CommentResponse.from(commentService.addComment(taskId, userId, request.body()));
    }

    @GetMapping("/{taskId}")
    public List<CommentResponse> getComments(@PathVariable String taskId) {
        return commentService.getCommentsByTask(taskId).stream().map(CommentResponse::from).toList();
    }
}
