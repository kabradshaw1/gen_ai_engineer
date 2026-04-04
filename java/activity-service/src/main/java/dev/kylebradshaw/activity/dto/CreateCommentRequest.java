package dev.kylebradshaw.activity.dto;

import jakarta.validation.constraints.NotBlank;

public record CreateCommentRequest(@NotBlank String body) {}
