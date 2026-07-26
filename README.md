# AI Customer Feedback Analyzer

An AI-powered customer feedback analysis service built with Go and LangChainGo.

The service classifies customer feedback, retrieves relevant business knowledge through tool calling, and generates a structured JSON report with deterministic validation.

## Architecture

```
Customer Feedback
        │
        ▼
 +----------------+
 |  Classifier    |
 +----------------+
        │
        ▼
 +----------------+
 |     Agent      |
 | (Tool Calling) |
 +----------------+
   │      │      │
   ▼      ▼      ▼
Customer Policy Workflow
 Tool     Tool     Tool
        │
        ▼
 +----------------+
 |    Reporter    |
 +----------------+
        │
        ▼
 Structured Report
```

## Key Features

- AI-powered feedback classification
- Tool-based knowledge retrieval
- Structured JSON report generation
- Deterministic report validation
- Human review detection
- Policy and workflow referencing
- Strict JSON schema enforcement

## Design Decisions

The system separates the analysis into three independent stages:

- **Classifier** determines the feedback category.
- **Agent** retrieves business knowledge using tool calling.
- **Reporter** validates, normalizes, and finalizes the response.

This separation keeps prompts simple, reduces hallucination, and guarantees a predictable output schema.

## Tech Stack

- Go
- LangChainGo
- OpenAI
- Gin
- Docker

## API

### Analyze Feedback

```http
POST /api/v1/feedback/analyze
```

### Health Check

```http
GET /health
```

## Running

Install dependencies

```bash
go mod tidy
```

Configure environment

```bash
touch .env
```

Set your OpenAI API key

```text
OPENAI_API_KEY=<your-api-key>
```

Run the application

```bash
go run main.go
```

## Example

### Request

1. Duplicate Charge (Billing)
```json
{
    "customer_id": "CUS001",
    "email": "unknown@example.com",
    "message": "I was charged twice for the same purchase.",
    "channel": "chat"
}
```

### Response

```json
{
    "feedback_id": "FB-1785085614572",
    "report": {
        "summary": "The customer claims they were charged twice for the same purchase. A customer profile matching CUS001 was found. It is unknown whether two successful charges exist for the same purchase and whether the customer is eligible for a refund.",
        "category": "billing",
        "urgency": "medium",
        "customer_context": {
            "available": true,
            "summary": "Customer profile was found."
        },
        "customer_claims": [
            "I was charged twice for the same purchase."
        ],
        "verified_facts": [
            "A customer profile matching CUS001 was found."
        ],
        "unknowns": [
            "Whether two successful charges exist for the same purchase",
            "Transaction identifiers and charge dates",
            "Charge amounts and payment method",
            "Refund eligibility after transaction verification"
        ],
        "references": [
            {
                "type": "policy",
                "id": "POL001",
                "title": "Refund Policy"
            },
            {
                "type": "workflow",
                "id": "WF001",
                "title": "Billing Complaint Workflow"
            }
        ],
        "recommended_next_steps": [
            "Verify the customer's identity and account ownership.",
            "Check transaction records for two successful charges related to the same purchase.",
            "Evaluate refund eligibility only after confirming the duplicate charge."
        ],
        "conditional_actions": [
            {
                "condition": "If two successful charges for the same purchase are verified and refund eligibility conditions are met",
                "action": "Submit a refund request for human approval",
                "requires_human_approval": true
            }
        ],
        "confidence": 0.95,
        "human_review_required": true,
        "human_review_reasons": [
            "restricted_action"
        ]
    }
}
```

2. Account Access Issue
```json
{
    "customer_id": "CUS001",
    "email": "unknown@example.com",
    "message": "I cannot log into my account even after resetting my password.",
    "channel": "chat"
}
```

### Response

```json
{
    "feedback_id": "FB-1785085676035",
    "report": {
        "summary": "The customer reports being unable to log into their account despite resetting the password, indicating a potential technical bug affecting access. The customer profile was found.",
        "category": "bug_report",
        "urgency": "medium",
        "customer_context": {
            "available": true,
            "summary": "Customer profile was found."
        },
        "customer_claims": [
            "I cannot log into my account even after resetting my password."
        ],
        "verified_facts": [
            "The feedback was submitted through the chat channel.",
            "A customer profile matching CUS001 was found."
        ],
        "unknowns": [
            "Error message or symptom details during login attempt",
            "Whether the password reset was successful and confirmed",
            "Steps taken after password reset",
            "Whether the issue is reproducible by support",
            "Affected product or feature details"
        ],
        "references": [
            {
                "type": "workflow",
                "id": "WF002",
                "title": "Bug Report Workflow"
            },
            {
                "type": "policy",
                "id": "POL002",
                "title": "Bug Handling Policy"
            }
        ],
        "recommended_next_steps": [
            "Verify the customer's identity and account ownership",
            "Collect detailed error messages and login attempt logs from the customer",
            "Attempt to reproduce the login issue in a controlled environment"
        ],
        "conditional_actions": [
            {
                "condition": "If the login issue is verified and confirmed to be a technical bug affecting account access",
                "action": "Submit an engineering ticket request for human approval",
                "requires_human_approval": true
            }
        ],
        "confidence": 0.85,
        "human_review_required": false,
        "human_review_reasons": []
    }
}
```

3. Product Bug Report
```json
{
    "customer_id": "CUS002",
    "email": "unknown@example.com",
    "message": "The mobile app crashes every time I try to upload a photo.",
    "channel": "chat"
}
```

### Response

```json
{
    "feedback_id": "FB-1785085713970",
    "report": {
        "summary": "The customer reports that the mobile app crashes every time they try to upload a photo, indicating a reproducible software bug with significant impact. The customer's identity and profile have been verified. It remains unknown what specific error messages or logs are generated and whether the issue can be reproduced internally.",
        "category": "bug_report",
        "urgency": "high",
        "customer_context": {
            "available": true,
            "summary": "Customer profile was found."
        },
        "customer_claims": [
            "The mobile app crashes every time I try to upload a photo."
        ],
        "verified_facts": [
            "Customer profile was found for CUS002.",
            "The feedback was submitted through the chat channel."
        ],
        "unknowns": [
            "Error messages or logs generated during the crash",
            "Steps to reproduce the crash internally",
            "App version and device details"
        ],
        "references": [
            {
                "type": "policy",
                "id": "POL002",
                "title": "Bug Handling Policy"
            },
            {
                "type": "workflow",
                "id": "WF002",
                "title": "Bug Report Workflow"
            }
        ],
        "recommended_next_steps": [
            "Verify the app version and device details from the customer.",
            "Collect logs and error messages related to the crash.",
            "Attempt to reproduce the crash internally using the provided details."
        ],
        "conditional_actions": [
            {
                "condition": "If the crash is reproduced and logs confirm a critical bug",
                "action": "Submit an engineering ticket request for human approval",
                "requires_human_approval": true
            }
        ],
        "confidence": 0.95,
        "human_review_required": false,
        "human_review_reasons": []
    }
}
```

4. Feature Request
```json
{
    "customer_id": "CUS001",
    "email": "unknown@example.com",
    "message": "It would be great if the dashboard supported dark mode.",
    "channel": "chat"
}
```

### Response

```json
{
    "feedback_id": "FB-1785085759870",
    "report": {
        "summary": "The customer requests that the dashboard support dark mode as a new feature. The customer profile was found.",
        "category": "feature_request",
        "urgency": "low",
        "customer_context": {
            "available": true,
            "summary": "Customer profile was found."
        },
        "customer_claims": [
            "It would be great if the dashboard supported dark mode."
        ],
        "verified_facts": [
            "The feedback was submitted through the chat channel.",
            "The customer profile was found."
        ],
        "unknowns": [
            "Business justification for the dark mode feature",
            "Customer impact details for the dark mode feature"
        ],
        "references": [
            {
                "type": "policy",
                "id": "FEATURE_POLICY_V1",
                "title": "Feature Request Policy"
            },
            {
                "type": "workflow",
                "id": "WF003",
                "title": "Feature Request Workflow"
            }
        ],
        "recommended_next_steps": [
            "Collect the business justification for the dark mode feature from the customer",
            "Record the customer impact and use case for the dark mode feature",
            "Submit the feature request to the Product Team following the standard workflow"
        ],
        "conditional_actions": [],
        "confidence": 0.95,
        "human_review_required": false,
        "human_review_reasons": []
    }
}
```

5. Ambiguous Feedback
```json
{
    "customer_id": "CUS002",
    "email": "unknown@example.com",
    "message": "It doesn't work.",
    "channel": "chat"
}
```

### Response

```json
{
    "feedback_id": "FB-1785085802677",
    "report": {
        "summary": "The customer states that the product or service 'doesn't work' without specifying any details about the issue or its impact. The feedback is vague and lacks information about the affected product, expected versus actual behavior, error messages, or steps to reproduce the problem.",
        "category": "unknown",
        "urgency": "low",
        "customer_context": {
            "available": true,
            "summary": "Customer profile was found."
        },
        "customer_claims": [
            "It doesn't work."
        ],
        "verified_facts": [
            "The feedback was submitted through the chat channel.",
            "A customer profile was found for CUS002."
        ],
        "unknowns": [
            "Affected product or feature",
            "Expected behavior and actual behavior",
            "Error message or observed symptom",
            "Steps leading to the issue",
            "Customer impact"
        ],
        "references": [],
        "recommended_next_steps": [
            "Request clarification from the customer about the specific product or feature that is not working",
            "Ask the customer to describe the expected behavior and the actual behavior observed",
            "Collect any error messages or symptoms and steps to reproduce the issue"
        ],
        "conditional_actions": [],
        "confidence": 0.2,
        "human_review_required": true,
        "human_review_reasons": [
            "ambiguous_category",
            "low_confidence",
            "missing_information"
        ]
    }
}
```

## Future Improvements

- Multi-agent orchestration
- Vector-based knowledge retrieval
- Conversation memory
- OpenTelemetry tracing
- Automated evaluation pipeline

## Design Write-up

### Why I structured the agent this way

I chose a **single-agent architecture with a fixed pipeline**:

`Classifier → Tool-calling Agent → Reporter`

The classifier performs a lightweight categorization, the agent retrieves business knowledge through tools, and the reporter deterministically validates and normalizes the final JSON. This separation keeps prompts simple, reduces hallucination, and guarantees a predictable output schema. A multi-agent architecture would add unnecessary complexity for the scope of this assignment.

### Production Improvements

For production, I would improve reliability with retries, timeouts, structured logging, and monitoring. To reduce cost and latency, I would cache policy/workflow lookups, use a smaller model for classification, and optimize prompts. I would also build an evaluation dataset to continuously measure classification accuracy, tool selection, schema validity, latency, and token usage.

### Deliberately Left Out

I intentionally omitted multi-agent orchestration, RAG/vector databases, persistent storage, authentication, and observability. These features are valuable in production but add infrastructure complexity without improving the core objective of this assignment: demonstrating structured LLM reasoning, tool calling, and deterministic report generation.

## License

This project was developed as part of a technical assessment.