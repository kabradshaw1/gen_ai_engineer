package dev.kylebradshaw.gateway.dto;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

@JsonIgnoreProperties(ignoreUnknown = true)
public class ActivityEventDto {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private String id;
    private String projectId;
    private String taskId;
    private String actorId;
    private String eventType;
    @JsonProperty("metadata")
    private Object rawMetadata;
    private String timestamp;

    public String getId() {
        return id;
    }

    public String getProjectId() {
        return projectId;
    }

    public String getTaskId() {
        return taskId;
    }

    public String getActorId() {
        return actorId;
    }

    public String getEventType() {
        return eventType;
    }

    public String getTimestamp() {
        return timestamp;
    }

    public String getMetadata() {
        if (rawMetadata == null) {
            return null;
        }
        if (rawMetadata instanceof String s) {
            return s;
        }
        try {
            return MAPPER.writeValueAsString(rawMetadata);
        } catch (JsonProcessingException e) {
            return rawMetadata.toString();
        }
    }
}
