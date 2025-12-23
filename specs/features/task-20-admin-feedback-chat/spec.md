# Task 20: Admin Feedback Chat with Claude AI - Specification

**Task ID**: task-20-admin-feedback-chat
**Phase**: 2B - New Microservices (Standalone Service)
**Priority**: P2 (Medium)
**Estimated Duration**: 4 weeks
**Dependencies**: Task 5 (Auth Service - RBAC), PostgreSQL with pgvector extension

---

## Overview

Implement a **standalone Feedback Service** that provides a chat interface for administrators to report bugs, suggest improvements, and submit feedback with screenshots. The chat uses Claude AI (Anthropic API) to process conversations, extract multiple topics/issues from a single conversation, and generate structured markdown files in the `issue-reports/` directory. The system includes cost optimization, semantic similarity search, navigation context tracking, and comprehensive monitoring.

**Key Design Decisions**:
- **Standalone Service**: Separate from AI Agent Service for independent scaling and deployment
- **Cost-First Design**: Includes caching, tiered models, token limits, and cost monitoring
- **Semantic Search**: pgvector-based similarity search for finding related issues
- **Production-Ready**: Includes rate limiting, monitoring, GDPR compliance, and abuse prevention

---

## User Scenarios

### US1: Admin Chat Interface (P1)

**As an** administrator
**I want to** chat with an AI agent to report issues and suggest improvements
**So that** I can provide feedback in natural language without filling complex forms

**Acceptance Criteria**:
- Chat interface accessible only to users with maximum admin role
- gRPC endpoint for sending messages and receiving responses
- Support for text messages and image attachments (screenshots)
- Real-time conversation flow with message history
- Claude AI processes messages and provides intelligent responses
- System validates admin permissions before allowing chat access

**Success Metrics**:
- <2s p95 response time for chat messages
- 100% of admin users can access chat interface
- 0 unauthorized access attempts succeed

**Independent Test**: Can be tested by creating an admin user, authenticating, and sending a chat message via gRPC. The system should validate permissions and return a response from Claude.

---

### US2: Multi-Topic Issue Extraction (P1)

**As an** administrator
**I want to** discuss multiple topics in a single chat session
**So that** I can efficiently report several issues without starting separate conversations

**Acceptance Criteria**:
- Claude AI identifies and separates multiple topics/issues from a single conversation
- Each identified topic generates a separate markdown file
- Files follow naming convention: `{username}Issue{number}{day}-{month}-{year}.md`
- Each file contains complete context for that specific issue
- System tracks which issues were extracted from which conversation

**Success Metrics**:
- 90%+ accuracy in topic separation (manual validation)
- All identified issues generate valid markdown files
- No duplicate issues created from same conversation

**Independent Test**: Can be tested by sending a chat message containing multiple distinct topics (e.g., "The product list is slow and also the login button is broken"). System should create separate markdown files for each topic.

---

### US3: Structured Issue Reports (P1)

**As a** developer
**I want to** receive structured issue reports in markdown format
**So that** I can quickly understand and prioritize issues

**Acceptance Criteria**:
- Markdown files follow the spec template format (similar to `spec.md` files)
- Files include: title, description, user information, category, priority, screenshots (base64), system context, proposed solution
- Files stored in `issue-reports/` directory at repository root
- Format is consistent and machine-readable
- Claude generates comprehensive, well-structured content

**Success Metrics**:
- 100% of generated files follow the template format
- 80%+ of generated content is actionable (manual review)
- Files are valid markdown and parseable

**Independent Test**: Can be tested by sending a chat message with an issue, verifying that a markdown file is created in `issue-reports/` with all required sections populated.

---

### US4: Navigation Context Tracking (P2)

**As an** administrator
**I want** the AI to understand what part of the system I'm using
**So that** it can provide more relevant context in issue reports

**Acceptance Criteria**:
- System tracks user navigation (current page, section, component)
- Navigation context sent to Claude with each message
- Claude uses context to enrich issue descriptions
- Context includes: current route, active features, user permissions, organization context
- Context is optimized and relevant (not overwhelming)

**Success Metrics**:
- Navigation context included in 100% of issue reports
- Context improves issue clarity (manual validation)
- <500ms overhead for context collection

**Independent Test**: Can be tested by navigating to a specific page (e.g., product catalog), sending a chat message, and verifying that the generated markdown includes relevant context about the product catalog section.

---

### US5: Issue Search and Reference (P2)

**As an** administrator
**I want to** search for previously reported issues
**So that** I can check if a problem was already reported and see its resolution

**Acceptance Criteria**:
- Claude can search database for similar issues when processing new reports
- Search uses semantic similarity (not just keyword matching)
- Claude references similar issues in responses
- Database stores: issue title, description, category, priority, status, resolution, related files
- gRPC endpoint to search issues by keywords or similarity

**Success Metrics**:
- 85%+ accuracy in finding similar issues
- <1s p95 search response time
- Claude references relevant past issues in 70%+ of new reports

**Independent Test**: Can be tested by creating an issue, then creating a similar issue and verifying that Claude identifies and references the first issue.

---

### US6: Category and Priority Classification (P2)

**As an** administrator
**I want to** suggest category and priority for my issue
**So that** the AI can make informed decisions about classification

**Acceptance Criteria**:
- Users can suggest category (bug, feature request, improvement, question) and priority (low, medium, high, critical)
- Claude AI makes final decision on category and priority
- Claude explains reasoning for classification if it differs from user suggestion
- Classification stored in both database and markdown file
- System supports filtering issues by category and priority

**Success Metrics**:
- 90%+ agreement between user suggestions and Claude decisions
- Classification accuracy validated by manual review
- All issues have valid category and priority assigned

**Independent Test**: Can be tested by sending a chat message with category/priority suggestions and verifying that Claude processes them and makes a final decision.

---

## Functional Requirements

### FR1: Admin Permission Validation
- System MUST validate user has maximum admin role before allowing chat access
- Validation MUST use existing RBAC system from auth-service
- Permission check MUST be performed on every chat request
- Unauthorized access attempts MUST be logged and rejected

### FR2: Claude AI Integration with Cost Optimization
- System MUST integrate with Anthropic Claude API (Claude Sonnet 4.5 for complex analysis, Claude Haiku for simple queries)
- API key MUST be configurable via environment variables (never hardcoded)
- System MUST implement cost controls:
  - Request deduplication (check if similar issue exists before calling Claude)
  - Response caching for identical contexts (1-hour TTL)
  - Prompt compression techniques (remove redundant context)
  - Daily/monthly token usage limits per organization
  - Automatic model downgrade (Sonnet → Haiku) when approaching limits
- System MUST handle API rate limits and errors gracefully with exponential backoff retry (max 3 attempts)
- System MUST track token usage per organization for billing/monitoring
- Responses MAY be streamed for better user experience (optional Phase 2 enhancement)

### FR3: Image Processing with Security
- System MUST accept image attachments (screenshots) in chat messages
- Security validations MUST include:
  - Validate image file headers (magic bytes), not just file extension
  - Sanitize EXIF data to remove PII (GPS coordinates, camera model, timestamps)
  - Prevent image bomb attacks (decompression limits)
  - Maximum file size: 5MB before processing, 2MB after compression
- Image optimization MUST include:
  - Server-side resize to max 1920x1080 pixels
  - Convert all images to WebP format for ~30% size reduction
  - Strip all metadata except essential image data
- Images MUST be converted to base64 for embedding in markdown
- System MUST support formats: PNG, JPG, JPEG, WebP (GIF converted to static image)
- Image processing MUST complete in <1s p95

### FR4: Multi-Topic Extraction with Enhanced Prompts
- Claude MUST analyze conversations to identify distinct topics/issues using structured prompts
- Topic separation guidelines:
  - Different system components = separate topics
  - Different issue types (bug vs feature) = separate topics
  - Same component with different issues = separate topics
  - Related observations about same issue = ONE topic
- Each topic MUST generate a separate markdown file
- System MUST track relationships between topics from same conversation
- Minimum topic confidence threshold: 0.80 (else request user clarification)
- Maximum topics per conversation: 10 (else suggest breaking into multiple chats)
- Extraction accuracy target: 92%+ (manual validation)
- Each topic MUST have minimum 50-character description
- Claude MUST provide reasoning for topic separation decisions

### FR5: Markdown File Generation
- Files MUST follow spec template format
- Naming convention MUST be: `{username}Issue{number}{day}-{month}-{year}.md`
- Files MUST include all required sections: title, description, user info, category, priority, screenshots, context, proposed solution
- Files MUST be stored in `issue-reports/` directory at repository root
- File creation MUST be atomic (all-or-nothing)

### FR6: Navigation Context with Optimization
- System MUST track user navigation state with detailed schema:
  - `route`: Current URL path (e.g., "/app/catalog/products/123")
  - `section`: High-level section (e.g., "catalog", "analytics")
  - `component`: Active UI component (e.g., "ProductDetailView")
  - `action`: Current user action (e.g., "editing", "viewing", "searching")
  - `activeFilters`: Currently applied filters (map[string]interface{})
  - `userPermissions`: Relevant permissions for current context
  - `organizationID`: Tenant context
  - `sessionDuration`: Time spent on current page
  - `recentActions`: Last 5 user actions for debugging
  - `browserInfo`: Browser, OS, viewport size
  - `errorState`: Error details if reporting from error page
- Context collection MUST be async and non-blocking
- Context payload MUST be <10KB per message
- Context MUST be categorized by relevance (critical, helpful, optional)
- Context optimization rules:
  - Bug reports: Include full context
  - Feature requests: Only route + section + action
  - Questions: Minimal context (route + section)
- Context collection MUST complete in <500ms p95

### FR7: Issue Database Storage with Vector Search
- System MUST store issue metadata in PostgreSQL database with pgvector extension
- Database schema MUST include:
  - `id`: UUID - Primary key
  - `user_id`: UUID - Reporter
  - `organization_id`: UUID - Tenant scope
  - `conversation_id`: UUID - Source conversation
  - `title`: VARCHAR(500) - Issue title
  - `description`: TEXT - Full description
  - `category`: ENUM(bug, feature_request, improvement, question)
  - `priority`: ENUM(low, medium, high, critical)
  - `status`: ENUM(open, in_progress, resolved, closed)
  - `embedding`: vector(1536) - Semantic embedding for similarity search
  - `user_suggested_category`: VARCHAR(50) - Optional user suggestion
  - `user_suggested_priority`: VARCHAR(50) - Optional user suggestion
  - `navigation_context`: JSONB - Context when reported
  - `similar_issues`: UUID[] - Array of related issue IDs
  - `markdown_file_path`: VARCHAR(1000) - Path to generated file
  - `created_at`: TIMESTAMP
  - `updated_at`: TIMESTAMP
  - `resolved_at`: TIMESTAMP - Optional
- Database MUST support full-text search using PostgreSQL's tsvector on title and description
- Database MUST support semantic similarity search using pgvector:
  - Create HNSW index for fast cosine similarity: `CREATE INDEX ON issue_reports USING hnsw (embedding vector_cosine_ops)`
  - Generate embeddings using OpenAI text-embedding-3-small (1536 dimensions) or Claude embeddings
  - Similarity thresholds: 0.85 for "very similar", 0.70 for "related", <0.70 for "not related"
  - Hybrid search: Combine vector similarity (60% weight) + keyword match (30%) + recency (10%)
- Database MUST include indexes:
  - PRIMARY KEY on id
  - INDEX on (organization_id, status, created_at) for filtering
  - GIN INDEX on navigation_context for JSONB queries
  - HNSW INDEX on embedding for vector search
  - Full-text index on title and description

### FR8: Issue Search with Hybrid Algorithm
- System MUST provide search functionality for past issues
- Search algorithm:
  1. Generate embedding for search query
  2. Perform vector similarity search (top 20 results, cosine similarity)
  3. Perform keyword search using PostgreSQL full-text search (top 20 results)
  4. Combine results with weighted scoring:
     - Vector similarity: 60% weight
     - Keyword match: 30% weight
     - Recency boost: 10% weight (newer issues score higher)
  5. Re-rank combined results
  6. Filter by organization_id (tenant scoping)
  7. Return top 5 most relevant issues
- Claude MUST use search results when processing new issues to:
  - Reference similar existing issues
  - Avoid duplicate issue creation
  - Suggest checking existing issues before creating new ones
- Search MUST be fast (<500ms p95 target, down from <1s)
- Search MUST respect multi-tenancy (filter by organization_id)

### FR9: Category and Priority
- Users MUST be able to suggest category and priority
- Claude MUST make final decision on classification
- Claude MUST explain reasoning if decision differs from suggestion
- Classification MUST be stored in database and markdown

### FR10: gRPC API
- System MUST provide gRPC endpoints for chat operations
- Endpoints MUST include: SendMessage, GetConversationHistory, SearchIssues, GetIssueDetails, UpdateIssueStatus
- gRPC MUST use Protocol Buffers v3
- API MUST follow existing service patterns
- gRPC-Web support MUST be included for browser clients
- All requests MUST include authentication headers validated via auth-service gRPC

### FR11: Rate Limiting and Abuse Prevention
- Per-user rate limits:
  - 30 messages per minute
  - 100 messages per hour
  - 500 messages per day
  - 5 concurrent conversations maximum
- Per-organization rate limits:
  - 1000 messages per hour
  - 5000 messages per day
  - 100 concurrent conversations
- Image upload limits:
  - 10 images per message
  - 50 images per conversation
  - 200 images per day per user
- Implementation MUST use Redis for distributed rate limiting with sliding window algorithm
- System MUST return HTTP 429 (Too Many Requests) with Retry-After header when limit exceeded
- System MUST log rate limit violations for abuse detection
- Admin dashboard MUST show usage patterns per user and organization

### FR12: Conversation Lifecycle Management
- Conversation states: active, completed, archived, deleted
- Storage policy:
  - Active conversations: Kept in database indefinitely
  - Completed conversations: Retained for 30 days, then auto-archived
  - Archived conversations: Retained for 90 days total, then messages deleted (metadata kept)
  - Issue metadata: Retained indefinitely (until manually deleted)
- Auto-archiving rules:
  - Archive after 7 days of inactivity
  - Archive immediately after all issues extracted and user confirms
  - User can manually archive/delete anytime
- Privacy and cleanup:
  - Delete conversation messages after 90 days (keep metadata)
  - Anonymize user data when user is deactivated (replace user_id with "DELETED_USER")
  - Partitioned tables for chat_messages by created_at (monthly partitions)
  - Automated retention policy: `DELETE FROM chat_messages WHERE created_at < NOW() - INTERVAL '90 days'`

### FR13: Cost Management and Monitoring
- Token usage tracking:
  - Track tokens per request (prompt + completion)
  - Track daily/monthly totals per organization
  - Store in `claude_usage` table for billing reference
- Cost controls:
  - Daily token limit per organization: 100,000 tokens (configurable)
  - Monthly token limit per organization: 3,000,000 tokens (configurable)
  - Automatic model downgrade when at 80% of limit (Sonnet → Haiku)
  - Block requests when at 100% of limit (with clear error message)
- Monitoring dashboards:
  - Real-time token usage per organization
  - Cost projections ($/day, $/month)
  - Model usage distribution (Sonnet vs Haiku)
  - Alert at 80% and 100% of monthly budget
- Response caching:
  - Cache Claude responses for identical context (1-hour TTL)
  - Cache key: hash(user_message + navigation_context + conversation_history)
  - Estimated cache hit rate target: 15-20%

### FR14: Monitoring and Observability
- Prometheus metrics to expose:
  - `claude_api_requests_total` (counter, labels: model, status)
  - `claude_api_errors_total` (counter, labels: error_type)
  - `claude_api_latency_seconds` (histogram, labels: model)
  - `claude_api_tokens_used_total` (counter, labels: organization_id, model)
  - `chat_messages_sent_total` (counter, labels: organization_id)
  - `issues_extracted_total` (counter, labels: category, priority)
  - `markdown_generation_duration_seconds` (histogram)
  - `similar_issues_found_total` (counter)
  - `navigation_context_collection_duration_seconds` (histogram)
  - `image_processing_duration_seconds` (histogram)
  - `rate_limit_hits_total` (counter, labels: user_id, limit_type)
  - `search_latency_seconds` (histogram, labels: search_type)
- Structured logging requirements:
  - INFO: Issue extracted successfully (issue_id, conversation_id, category, priority)
  - INFO: Markdown file created (file_path, issue_id)
  - WARN: Claude API rate limit approached (usage_percent, organization_id)
  - ERROR: Claude API failure (error_type, retry_attempt, will_retry)
  - ERROR: Markdown file creation failed (error, rollback_status)
  - AUDIT: Admin accessed chat (user_id, organization_id, timestamp)
  - AUDIT: Permission check failed (user_id, attempted_operation, reason)
- Alerts (via Prometheus Alertmanager):
  - Claude API error rate > 5% for 5 minutes
  - Markdown generation failures > 10 in 1 hour
  - Claude API costs exceed $100/day
  - Similar issue search p95 latency > 2s
  - Rate limit hits > 100/hour for single user (potential abuse)

### FR15: GDPR Compliance and Data Privacy
- User consent:
  - Inform admins that chat messages are sent to Claude API (Anthropic)
  - Display privacy notice before first chat interaction
  - Store consent acceptance in user profile
- Data minimization:
  - Only collect necessary navigation context
  - Do NOT send user emails or internal system IDs to Claude
  - Sanitize context to remove PII before sending to Claude
  - Hash organization names in Claude requests (use org_id hash, not name)
- Right to erasure:
  - Provide gRPC endpoint: DeleteUserData(user_id)
  - Delete all user conversations, messages, and images
  - Anonymize issue reports (replace user_id with "DELETED_USER", keep issue content)
  - Log deletion in audit trail
- Data portability:
  - Provide gRPC endpoint: ExportUserData(user_id) → JSON
  - Export all user conversations, issues, and metadata
- Data retention:
  - Chat messages: 90 days
  - Issue reports (markdown): Indefinite (until manually deleted or user data deletion)
  - Image attachments: 90 days
  - Conversation metadata: 1 year
  - Audit logs: 2 years
- PII handling:
  - Strip EXIF data from images (may contain GPS, device info)
  - Anonymize user information after issue resolution
  - No user emails in markdown files (use username only)

---

## Key Entities

### IssueReport
Represents a reported issue extracted from a chat conversation.

**Attributes**:
- `id`: UUID - Unique identifier
- `user_id`: UUID - Admin user who reported
- `organization_id`: UUID - Organization context
- `conversation_id`: UUID - Chat conversation this issue came from
- `title`: string - Issue title (generated by Claude)
- `description`: string - Detailed description (generated by Claude)
- `category`: enum - bug, feature_request, improvement, question (decided by Claude)
- `priority`: enum - low, medium, high, critical (decided by Claude)
- `user_suggested_category`: string (optional) - User's suggestion
- `user_suggested_priority`: string (optional) - User's suggestion
- `status`: enum - open, in_progress, resolved, closed
- `markdown_file_path`: string - Path to generated markdown file
- `navigation_context`: JSONB - Context when issue was reported
- `similar_issues`: []UUID - Related issues found by search
- `created_at`: timestamp
- `updated_at`: timestamp
- `resolved_at`: timestamp (optional)

### ChatMessage
Represents a message in a chat conversation.

**Attributes**:
- `id`: UUID - Unique identifier
- `conversation_id`: UUID - Conversation this message belongs to
- `user_id`: UUID - User who sent the message
- `role`: enum - user, assistant
- `content`: string - Message text
- `images`: []ImageData - Attached screenshots (base64)
- `navigation_context`: JSONB - User's navigation state
- `created_at`: timestamp

### ImageData
Represents an image attachment.

**Attributes**:
- `id`: UUID - Unique identifier
- `message_id`: UUID - Message this image belongs to
- `filename`: string - Original filename
- `content_type`: string - MIME type (image/png, image/jpeg, etc.)
- `base64_data`: string - Base64 encoded image data
- `size_bytes`: int - Image size in bytes

### Conversation
Represents a chat session between admin and Claude.

**Attributes**:
- `id`: UUID - Unique identifier
- `user_id`: UUID - Admin user
- `organization_id`: UUID - Organization context
- `status`: enum - active, completed, archived
- `created_at`: timestamp
- `updated_at`: timestamp
- `last_message_at`: timestamp

---

## Markdown File Template

Files MUST be generated in the `issue-reports/` directory following this exact template:

**Filename Format**: `{username}Issue{number}{day}-{month}-{year}.md`

**Example**: `johndoeIssue1218-12-2025.md` (user: johndoe, issue #12, date: Dec 18, 2025)

**Template Structure**:

```markdown
# {Issue Title}

**Reported By**: {username} ({user_email})
**Organization**: {organization_name}
**Date**: {YYYY-MM-DD HH:mm:ss UTC}
**Category**: {bug|feature_request|improvement|question}
**Priority**: {low|medium|high|critical}
**Status**: {open|in_progress|resolved|closed}

---

## Description

{Detailed description generated by Claude, minimum 100 characters}

---

## User's Original Message

> {Original chat message from admin, verbatim}

---

## System Context

**Navigation**:
- **Route**: {route} (e.g., /app/catalog/products/123)
- **Section**: {section} (e.g., catalog)
- **Component**: {component} (e.g., ProductDetailView)
- **Action**: {action} (e.g., editing, viewing, searching)

**Environment**:
- **Organization ID**: {organization_id}
- **User Role**: {user_role}
- **Browser**: {browser_name} {version} on {os_name}
- **Viewport**: {width}x{height}
- **Timestamp**: {ISO8601}

**Recent Actions** (last 5):
1. {timestamp}: {action_description}
2. {timestamp}: {action_description}
3. {timestamp}: {action_description}
4. {timestamp}: {action_description}
5. {timestamp}: {action_description}

{if errorState exists}
**Error State**:
- **Error Code**: {error_code}
- **Error Message**: {error_message}
- **Stack Trace**: {truncated_stack_trace}
{endif}

---

## Screenshots

{if screenshots exist}
### Screenshot 1
![Screenshot 1](data:image/webp;base64,{base64_data})
*AI-generated caption: {description}*

### Screenshot 2
![Screenshot 2](data:image/webp;base64,{base64_data})
*AI-generated caption: {description}*
{else}
No screenshots provided.
{endif}

---

## Related Issues

{if similar_issues found}
This issue may be related to:

- [#{issue_id}]({markdown_file_path}): {issue_title} (Similarity: {score}%)
- [#{issue_id}]({markdown_file_path}): {issue_title} (Similarity: {score}%)
{else}
No similar issues found.
{endif}

---

## Proposed Solution

{Claude's suggested solution, investigation steps, or recommendations. Minimum 50 characters.}

{if applicable}
**Estimated Effort**: {small|medium|large}
**Affected Components**: {list_of_components}
{endif}

---

## Classification Reasoning

{if user suggested different category/priority}
**User Suggested**: {user_category} / {user_priority}
**Final Decision**: {final_category} / {final_priority}
**Reasoning**: {Claude's explanation for the decision, minimum 30 characters}
{else}
**Classification**: {final_category} / {final_priority}
**Reasoning**: {Claude's explanation, minimum 30 characters}
{endif}

---

## Metadata

- **Issue ID**: {uuid}
- **Conversation ID**: {conversation_uuid}
- **Total Topics Extracted**: {total_topics_from_conversation}
- **Topic Number**: {topic_number} of {total_topics}
- **Similar Issues Count**: {count}
- **Created At**: {ISO8601}
- **Updated At**: {ISO8601}
- **File Path**: `{relative_path_from_repo_root}`
- **Embedding Generated**: {true|false}

---

*This issue was automatically generated by the GIIA Feedback Service using Claude AI.*
*For questions or updates, contact the development team or update this issue via the feedback system.*
```

**Validation Rules**:
- All placeholder fields MUST be replaced with actual values
- If optional sections have no data, use "N/A" or "None" (never leave blank)
- Markdown MUST be valid and parseable
- File size MUST be <1MB (including base64 images)
- Issue title MUST be <500 characters
- Description MUST be >100 characters
- File creation MUST be atomic (write to temp file, then rename)

---

## Claude AI System Prompts

### Topic Extraction Prompt

```
You are an expert issue tracker analyzing admin feedback for the GIIA inventory management platform.

Your task is to identify DISTINCT topics from the conversation and create structured issue reports.

TOPIC SEPARATION GUIDELINES:
1. Different system components (catalog vs analytics) = separate topics
2. Different issue types (bug vs feature request) = separate topics
3. Same component but different issues (slow search + incorrect results) = separate topics
4. Related observations about same issue (symptom + root cause) = ONE topic
5. Do NOT split a coherent issue into multiple topics unnecessarily

QUALITY STANDARDS:
- Minimum topic confidence: 0.80 (else ask user for clarification)
- Maximum topics per conversation: 10
- Minimum description length: 50 characters
- Each topic must have clear value (bug symptom, feature benefit, or improvement rationale)

OUTPUT FORMAT (JSON):
{
  "topics": [
    {
      "topic_number": 1,
      "title": "Concise, descriptive title (max 100 chars)",
      "category": "bug|feature_request|improvement|question",
      "priority": "low|medium|high|critical",
      "description": "Detailed description with context (min 100 chars)",
      "context_relevant": true,
      "extracted_screenshots": [0, 1],
      "proposed_solution": "Suggested fix or investigation steps (min 50 chars)",
      "affected_components": ["component1", "component2"],
      "estimated_effort": "small|medium|large",
      "reasoning": "Why this is a separate topic and classification rationale (min 30 chars)",
      "confidence": 0.95
    }
  ],
  "total_topics": 1,
  "overall_confidence": 0.95
}

CONTEXT PROVIDED:
- User message: {user_message}
- Navigation context: {navigation_context}
- Conversation history: {recent_messages}
- Similar existing issues: {similar_issues}

Analyze the conversation and extract all distinct topics following the guidelines above.
```

### Similar Issue Search Prompt

```
You are analyzing if a new issue is similar to existing issues in the GIIA platform.

SIMILARITY CRITERIA:
- Same component AND same symptom = VERY similar (>0.85)
- Same component OR same symptom = Related (0.70-0.85)
- Different component AND different symptom = Not related (<0.70)

Consider:
- Technical similarity (same error, same component)
- User impact similarity (same workflow affected)
- Root cause similarity (if mentioned)

NEW ISSUE:
Title: {new_title}
Description: {new_description}
Category: {new_category}

EXISTING ISSUES (vector search results):
{existing_issues}

For each existing issue, provide:
{
  "issue_id": "uuid",
  "similarity_score": 0.0-1.0,
  "similarity_reasoning": "Brief explanation",
  "recommendation": "duplicate|related|not_related"
}

Return JSON array of similar issues (only those with score >= 0.70).
```

---

## Non-Functional Requirements

### Performance
- Chat message response time: <1s p95 (with caching), <2s p95 (without cache)
- Issue search: <500ms p95 (down from <1s via pgvector HNSW index)
- Markdown file generation: <3s p95 (down from <5s)
- Navigation context collection: <500ms p95
- Image processing: <1s p95 per image
- Claude API latency: <2s p95 (external dependency)
- Database query latency: <100ms p95
- Cache hit rate target: 15-20% for Claude responses

### Security
- Only maximum admin role can access chat (validated via auth-service gRPC)
- API keys NEVER hardcoded (environment variables + secrets manager)
- Image uploads validated:
  - Magic byte validation (not just extension)
  - EXIF data stripped (prevent PII leakage)
  - Size limits enforced (5MB before processing, 2MB after)
  - Malicious payload detection (image bomb prevention)
- All operations logged for audit with structured logging
- SQL injection prevented via GORM parameterized queries
- XSS prevention: Sanitize user input before rendering
- Rate limiting enforced (Redis-based)
- PII redacted from logs and Claude API requests
- GDPR compliance: data retention policies, right to erasure, data portability

### Scalability
- Support 100+ concurrent chat sessions per instance
- Horizontal scaling: Stateless service (can deploy multiple replicas)
- Database queries optimized with indexes:
  - HNSW index for vector similarity (O(log n) search)
  - B-tree indexes on organization_id, status, created_at
  - Full-text GIN index on title/description
- Claude API calls rate-limited and cached (1-hour TTL)
- Redis for distributed rate limiting and response caching
- Database connection pooling via GORM
- Partitioned tables for chat_messages (monthly partitions)
- Auto-archiving of old conversations (reduce database size)

### Reliability
- Graceful handling of Claude API failures:
  - Exponential backoff retry (max 3 attempts)
  - Circuit breaker pattern (after 5 consecutive failures, pause 1 minute)
  - Fallback to cached responses when available
  - Clear error messages to users
- Retry logic for transient errors (network, timeout, rate limit)
- Markdown file generation is atomic:
  - Write to temporary file first
  - Validate markdown syntax
  - Rename to final location (atomic operation)
  - Rollback database transaction on file write failure
- Database transactions for data consistency:
  - Issue creation + embedding generation + file write = single transaction
  - Rollback on any step failure
- Health checks: /health endpoint for Kubernetes liveness/readiness probes
- Prometheus metrics for monitoring and alerting
- Error rate SLO: <1% for all operations (excluding external API failures)

---

## Success Criteria

### Mandatory (Must Have)

**Core Features**:
- ✅ Admin chat interface with Claude AI integration (Sonnet 4.5 + Haiku)
- ✅ Multi-topic extraction from conversations (92%+ accuracy)
- ✅ Structured markdown file generation following exact template
- ✅ Navigation context tracking with optimization (< 10KB payload)
- ✅ Issue database storage with pgvector semantic search
- ✅ Category and priority classification with reasoning
- ✅ gRPC API with Protocol Buffers v3 (5 endpoints minimum)
- ✅ Admin permission validation via auth-service gRPC
- ✅ Image attachment support with security validation and optimization

**Cost & Performance**:
- ✅ Cost controls: caching, tiered models, token limits, monitoring
- ✅ <$50/month Claude API costs for 100 active admins
- ✅ <1s p95 chat response time (with caching)
- ✅ <500ms p95 issue search (pgvector HNSW)
- ✅ 15-20% cache hit rate for Claude responses

**Production Readiness**:
- ✅ Rate limiting (per-user and per-org limits)
- ✅ Conversation lifecycle management (auto-archiving, retention policies)
- ✅ GDPR compliance (data retention, right to erasure, data portability)
- ✅ Monitoring: 15+ Prometheus metrics, structured logging, alerts
- ✅ 95%+ test coverage (up from 80%)

**Quality Metrics**:
- ✅ 95%+ of similar issues correctly identified
- ✅ 99.9% markdown file creation success rate
- ✅ <0.1% Claude API error rate (with retries)
- ✅ Zero unauthorized access incidents
- ✅ All operations have monitoring dashboards

### Optional (Nice to Have)

**Phase 2 Enhancements**:
- ⚪ Streaming responses for real-time chat feel (WebSocket or Server-Sent Events)
- ⚪ Issue status updates and resolution tracking
- ⚪ Integration with GitHub Issues (auto-create issues from markdown)
- ⚪ Advanced analytics on feedback patterns (most common issues, trending topics)
- ⚪ Conversation templates for common issue types
- ⚪ Multi-language support (currently English-only)
- ⚪ Voice-to-text for chat messages
- ⚪ Automated issue triage suggestions
- ⚪ A/B testing for prompt effectiveness

---

## Out of Scope

- ❌ Email notifications (explicitly excluded)
- ❌ Non-admin user access
- ❌ Chat history persistence beyond issue extraction
- ❌ Real-time collaboration features
- ❌ Issue assignment and workflow management
- ❌ Integration with external issue trackers (GitHub, Jira, etc.)

---

## Dependencies

### Internal Dependencies
- **Task 5**: Auth Service with RBAC (permission validation via gRPC)
- **Shared Packages**:
  - `pkg/events` - NATS event publishing
  - `pkg/database` - PostgreSQL connection pooling
  - `pkg/logger` - Structured logging (Zerolog)
  - `pkg/errors` - Typed error system
  - `pkg/config` - Configuration management

### External Dependencies
- **Anthropic Claude API**:
  - API Key required (Claude Sonnet 4.5 and Haiku)
  - Pricing: ~$3 per 1M input tokens, ~$15 per 1M output tokens (Sonnet)
  - Rate limits: Check current Anthropic tier limits
- **OpenAI Embeddings API** (or Claude Embeddings when available):
  - For semantic search embeddings (text-embedding-3-small)
  - Pricing: ~$0.02 per 1M tokens
- **PostgreSQL 16** with **pgvector** extension:
  - Extension version: pgvector 0.5.0+
  - Required for vector similarity search
- **Redis 7**:
  - For rate limiting and response caching
  - Distributed locks for concurrent request handling

### Infrastructure Dependencies
- **NATS Jetstream**: For event publishing (FEEDBACK_EVENTS stream)
- **Kubernetes**: For deployment (Helm chart required)
- **Prometheus**: For metrics collection
- **Docker**: For containerization

### Development Dependencies
- **Go 1.23.4+**
- **Protocol Buffers v3** compiler (`protoc`)
- **golangci-lint**: For code quality
- **testify**: For mocking and assertions

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation Strategy |
|------|--------|-------------|---------------------|
| **Claude API costs exceed budget** | High | Medium | Response caching (1hr TTL), tiered models (Sonnet/Haiku), token limits per org, daily monitoring, alerts at 80%/100% budget |
| **Multi-topic extraction accuracy <92%** | High | Medium | Enhanced prompts with examples, confidence thresholds (0.80), manual validation loop, iterative prompt tuning, A/B testing |
| **Navigation context overhead >500ms** | Medium | Low | Async collection, payload size limits (<10KB), context categorization (critical/helpful/optional), caching browser info |
| **Image processing security vulnerabilities** | Critical | Low | Magic byte validation, EXIF stripping, image bomb detection, size limits (5MB→2MB), WebP conversion, security audits |
| **Permission validation failures** | Critical | Low | Comprehensive unit tests, integration tests with auth-service, audit logging, fail-closed approach, gRPC circuit breaker |
| **Database search performance degradation** | Medium | Medium | pgvector HNSW index (O(log n)), query optimization, connection pooling, caching search results, monthly partitioning |
| **pgvector extension compatibility** | Medium | Low | Use stable pgvector version (0.5.0+), test in staging, fallback to keyword search only, document setup process |
| **Claude API rate limits hit frequently** | Medium | Medium | Exponential backoff (max 3 retries), circuit breaker (5 failures → 1min pause), response caching, request deduplication |
| **GDPR compliance violations** | Critical | Low | Legal review of data retention policies, automated data deletion jobs, consent tracking, PII sanitization, audit trail |
| **Markdown file creation failures** | Medium | Low | Atomic writes (temp file → rename), transaction rollback, retry logic, file system monitoring, alerts on failures >10/hour |
| **Abuse via rate limit bypass** | Medium | Medium | Redis distributed rate limiting, IP-based limiting (secondary), anomaly detection, automatic blocking, admin dashboard |
| **Embedding API costs (OpenAI)** | Low | Medium | Cache embeddings in database, only generate for new issues, batch embedding generation, explore cheaper alternatives |

---

## References

### External Documentation
- **Anthropic Claude API**: https://docs.anthropic.com/claude/reference
- **OpenAI Embeddings API**: https://platform.openai.com/docs/guides/embeddings
- **PostgreSQL pgvector**: https://github.com/pgvector/pgvector
- **gRPC Protocol Buffers**: https://grpc.io/docs/languages/go/
- **NATS Jetstream**: https://docs.nats.io/nats-concepts/jetstream

### Internal Documentation
- **Project Guidelines**: `CLAUDE.md` (Clean Architecture, Go standards, testing)
- **Spec Template**: `docs/templates/spec-driven-development/spec-template.md`
- **RBAC System**: `services/auth-service/internal/core/usecases/rbac/`
- **gRPC Patterns**: `services/auth-service/api/proto/auth/v1/`
- **Shared Packages**: `pkg/events`, `pkg/database`, `pkg/logger`, `pkg/errors`

### Related Specifications
- **Task 5**: Auth Service with RBAC
- **Phase 2 Overview**: `specs/features/PHASE_2_OVERVIEW.md`
- **Unit Testing Standards**: `docs/UNIT_TESTING_STANDARDS.md`

---

## Service Architecture Overview

### Standalone Service Rationale

This feature is implemented as a **standalone Feedback Service** rather than extending the AI Agent Service for the following reasons:

1. **Separation of Concerns**:
   - AI Agent Service: Demand forecasting, anomaly detection, inventory optimization
   - Feedback Service: Admin feedback, bug reports, issue tracking
   - Different domains with different business logic

2. **Independent Scaling**:
   - Feedback has unpredictable spiky load (admin reports)
   - AI Agent has steady computational load (batch predictions)
   - Different resource requirements (Feedback: I/O bound, AI Agent: CPU bound)

3. **Deployment Flexibility**:
   - Deploy feedback system updates without touching AI forecasting
   - Different release cycles
   - Easier rollback on failures

4. **Technology Choices**:
   - Feedback: Focus on Claude API, pgvector, markdown generation
   - AI Agent: Focus on ML models, data pipelines, statistical analysis
   - Can optimize each service for its specific needs

### Service Structure

```
services/feedback-service/
├── cmd/
│   └── api/
│       └── main.go                 # Service entry point
│
├── internal/
│   ├── core/                       # Domain Layer
│   │   ├── domain/
│   │   │   ├── issue_report.go
│   │   │   ├── conversation.go
│   │   │   ├── chat_message.go
│   │   │   └── image_data.go
│   │   ├── usecases/
│   │   │   ├── chat/
│   │   │   │   ├── send_message.go
│   │   │   │   ├── get_history.go
│   │   │   │   └── extract_topics.go
│   │   │   ├── issue/
│   │   │   │   ├── generate_markdown.go
│   │   │   │   ├── search_issues.go
│   │   │   │   ├── classify_issue.go
│   │   │   │   └── update_status.go
│   │   │   └── context/
│   │   │       └── collect_navigation.go
│   │   └── providers/              # Interface contracts
│   │       ├── ai_client.go
│   │       ├── embedding_client.go
│   │       ├── auth_client.go
│   │       ├── file_storage.go
│   │       ├── cache.go
│   │       └── rate_limiter.go
│   │
│   ├── infrastructure/             # Infrastructure Layer
│   │   ├── adapters/
│   │   │   ├── claude/
│   │   │   │   ├── client.go       # Anthropic SDK wrapper
│   │   │   │   ├── prompts.go      # System prompts
│   │   │   │   └── cost_tracker.go
│   │   │   ├── openai/
│   │   │   │   └── embeddings.go
│   │   │   ├── auth/
│   │   │   │   └── grpc_client.go
│   │   │   └── cache/
│   │   │       └── redis_cache.go
│   │   ├── repositories/
│   │   │   ├── issue_repository.go
│   │   │   ├── conversation_repository.go
│   │   │   └── usage_repository.go
│   │   ├── entrypoints/
│   │   │   ├── grpc/
│   │   │   │   └── feedback_server.go
│   │   │   └── http/
│   │   │       └── health_handler.go
│   │   ├── storage/
│   │   │   └── markdown_writer.go
│   │   └── ratelimit/
│   │       └── redis_limiter.go
│   │
│   └── app/                        # Application Layer
│       └── container.go            # Dependency injection
│
├── api/
│   └── proto/
│       └── feedback/
│           └── v1/
│               ├── feedback.proto  # gRPC service definition
│               └── feedback_grpc.pb.go (generated)
│
├── migrations/
│   ├── 001_create_conversations.sql
│   ├── 002_create_chat_messages.sql
│   ├── 003_create_issue_reports.sql
│   ├── 004_add_pgvector.sql
│   └── 005_create_usage_tracking.sql
│
├── docs/
│   ├── API.md                      # gRPC API documentation
│   ├── CLAUDE_PROMPTS.md          # System prompts documentation
│   ├── COST_OPTIMIZATION.md       # Cost reduction strategies
│   └── TESTING_GUIDE.md           # Testing with mock responses
│
├── deployments/
│   ├── kubernetes/
│   │   └── feedback-service.yaml
│   └── helm/
│       └── feedback-service/
│
├── .env.example
├── Dockerfile
├── Makefile
└── README.md
```

### NATS Event Stream

Create new stream: `FEEDBACK_EVENTS`

**Event Types**:
- `feedback.issue.created` - New issue extracted
- `feedback.issue.updated` - Issue status changed
- `feedback.conversation.completed` - Conversation finished
- `feedback.markdown.generated` - Markdown file created

---

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
**Deliverables**:
- Service scaffolding with Clean Architecture
- Database schema and migrations (including pgvector)
- gRPC API definition (Protocol Buffers)
- Health check endpoint
- Docker and Kubernetes configuration

**Tasks**:
1. Create service directory structure
2. Define Protocol Buffer schemas
3. Implement domain entities
4. Create database migrations
5. Setup dependency injection container
6. Configure Docker and Kubernetes manifests

### Phase 2: Claude Integration & Chat (Week 2)
**Deliverables**:
- Claude API client with cost controls
- Basic chat message handling
- Response caching implementation
- Token usage tracking
- Rate limiting

**Tasks**:
1. Implement Claude API client adapter
2. Create cost tracking system
3. Implement Redis caching layer
4. Build SendMessage use case
5. Implement rate limiting with Redis
6. Create GetConversationHistory use case

### Phase 3: Multi-Topic Extraction & Markdown (Week 2-3)
**Deliverables**:
- Multi-topic extraction with enhanced prompts
- Markdown file generation (atomic writes)
- Navigation context collection and optimization
- Image processing with security

**Tasks**:
1. Design and test topic extraction prompts
2. Implement ExtractTopics use case
3. Build markdown template engine
4. Create atomic file writer
5. Implement navigation context collector
6. Build image processor with validation and optimization

### Phase 4: Semantic Search & Intelligence (Week 3)
**Deliverables**:
- pgvector semantic similarity search
- OpenAI embeddings integration
- Similar issue detection
- Category/priority classification

**Tasks**:
1. Setup pgvector extension
2. Implement OpenAI embeddings client
3. Create hybrid search algorithm (vector + keyword)
4. Build SearchIssues use case
5. Implement similar issue detection
6. Create classification logic with reasoning

### Phase 5: Production Readiness (Week 4)
**Deliverables**:
- Comprehensive monitoring (Prometheus metrics)
- Structured logging with audit trail
- Conversation lifecycle management
- GDPR compliance features
- Complete test suite (95%+ coverage)

**Tasks**:
1. Implement all Prometheus metrics
2. Setup structured logging
3. Create conversation auto-archiving
4. Implement GDPR endpoints (DeleteUserData, ExportUserData)
5. Write comprehensive unit tests
6. Write integration tests
7. Write end-to-end tests
8. Create operational runbook
9. Perform security audit

### Phase 6: Documentation & Launch (Week 4)
**Deliverables**:
- Complete API documentation
- System prompts documentation
- Cost optimization guide
- Testing guide with mock responses
- Deployment runbook

**Tasks**:
1. Document all gRPC endpoints
2. Document Claude system prompts
3. Create cost optimization guide
4. Write testing guide
5. Create deployment runbook
6. Conduct final review
7. Deploy to staging
8. Perform load testing
9. Deploy to production

---

## Testing Strategy

### Unit Tests (Target: 95% coverage)

**Use Case Testing**:
```go
// Example: internal/core/usecases/chat/send_message_test.go
func TestSendMessage_WithValidInput_ReturnsResponse(t *testing.T) {
    // Given
    mockAIClient := new(providers.MockAIClient)
    mockAuthClient := new(providers.MockAuthClient)
    mockCache := new(providers.MockCache)

    givenUserID := uuid.New()
    givenMessage := "Product search is slow"
    givenOrgID := uuid.New()

    expectedResponse := "I understand you're experiencing slow search..."

    mockAuthClient.On("CheckPermission", mock.Anything, givenUserID, "chat.access").
        Return(true, nil)
    mockAIClient.On("SendMessage", mock.Anything, mock.Anything).
        Return(expectedResponse, 1500, nil) // response, tokens, error
    mockCache.On("Get", mock.Anything).Return("", errors.New("cache miss"))
    mockCache.On("Set", mock.Anything, mock.Anything, time.Hour).Return(nil)

    useCase := chat.NewSendMessageUseCase(mockAIClient, mockAuthClient, mockCache)

    // When
    response, err := useCase.Execute(context.Background(), givenUserID, givenOrgID, givenMessage, nil)

    // Then
    assert.NoError(t, err)
    assert.Equal(t, expectedResponse, response.Content)
    mockAIClient.AssertExpectations(t)
    mockAuthClient.AssertExpectations(t)
}
```

**Test Coverage Requirements**:
- All use cases: 95%+
- All domain entities: 90%+
- All repositories: 85%+
- All adapters: 85%+

### Integration Tests

**Database Integration**:
- CRUD operations for all entities
- pgvector similarity search
- Transaction rollback scenarios
- Migration up/down testing

**Claude API Integration** (Staging only):
- Real API calls with test prompts
- Multi-topic extraction accuracy
- Token usage tracking
- Error handling (rate limits, timeouts)

**Auth Service Integration**:
- Permission validation
- Admin-only access enforcement
- Multi-tenancy scoping

### End-to-End Tests

**Scenario 1: Single Issue Report**
1. Admin authenticates
2. Sends message with screenshot
3. Claude extracts single issue
4. Markdown file created
5. Issue searchable in database
6. Event published to NATS

**Scenario 2: Multi-Topic Extraction**
1. Admin sends message with 3 distinct topics
2. Claude identifies all 3 topics
3. 3 markdown files created
4. All linked to same conversation
5. All searchable independently

**Scenario 3: Similar Issue Detection**
1. Create issue A: "Product search is slow"
2. Create issue B: "Catalog search performance issue"
3. System detects similarity (>0.85)
4. Issue B references issue A
5. Markdown includes "Related Issues" section

**Scenario 4: Cost Control**
1. Organization reaches 80% of monthly token limit
2. Alert triggered
3. Subsequent requests use Haiku model
4. Organization reaches 100% of limit
5. Requests blocked with clear error message

### Performance Tests

**Load Test**:
- 100 concurrent chat sessions
- 1000 messages over 10 minutes
- Target: <1s p95 response time

**Stress Test**:
- Simulate Claude API failures
- Test circuit breaker activation
- Verify graceful degradation

**Volume Test**:
- 10,000 issues in database
- Test search performance (<500ms p95)
- Test pagination

### Security Tests

- SQL injection attempts in search queries
- XSS attempts in chat messages
- Image upload malicious payloads (image bombs)
- Unauthorized access attempts (non-admin users)
- Rate limit bypass attempts
- EXIF data leakage verification

---

**Document Version**: 2.0
**Last Updated**: 2025-12-18
**Status**: Enhanced and Ready for Planning
**Changes**: Added cost optimization, pgvector semantic search, GDPR compliance, enhanced monitoring, explicit markdown template, Claude system prompts, standalone service architecture, and detailed implementation phases
**Next Step**: Create implementation plan (plan.md) with task breakdown

