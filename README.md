# Maileroo Go SDK

[Maileroo](https://maileroo.com) is a robust email delivery platform designed for effortless sending of transactional and marketing emails. This Go SDK offers a straightforward interface for working with the Maileroo API, supporting basic
email formats, templates, bulk sending, and scheduling capabilities.

## Features

- Send basic HTML or plain text emails with ease
- Use pre-defined templates with dynamic data
- Send up to 500 personalized emails in bulk
- Schedule emails for future delivery
- Manage scheduled emails (list & delete)
- Add tags, custom headers, and reference IDs
- Attach files to your emails
- Support for multiple recipients, CC, BCC, and Reply-To
- Enable or disable open and click tracking
- Built-in input validation and error handling

## Installation

Install the SDK using the following command:

```bash
go get github.com/maileroo/maileroo-go-sdk
```

Then, include the SDK in your project:

```
import "github.com/maileroo/maileroo-go-sdk/maileroo"
```

## Quick Start

```
client, err := maileroo.NewClient("your-api-key", 30)

if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

htmlContent := "<h1>Test Email</h1><p>This is a test email from the Maileroo Go SDK.</p>"
plainContent := "Test Email\n\nThis is a test email from the Maileroo Go SDK."

referenceId, err := client.SendBasicEmail(context.Background(), maileroo.BasicEmailData{
    From: maileroo.NewEmail("YOUR_EMAIL_ADDRESS", "Your Name"),
    To: []maileroo.EmailAddress{
        maileroo.NewEmail("RECIPIENT_EMAIL_ADDRESS", "Recipient Name"),
    },
    Cc: []maileroo.EmailAddress{
        maileroo.NewEmail("CC_EMAIL_ADDRESS", "CC Name"),
    },
    Subject: "Test Email from Maileroo Go SDK",
    HTML:    &htmlContent,
    Plain:   &plainContent,
})

if err != nil {
    log.Fatalf("Failed to send email: %v", err)
}

log.Printf("Email sent with reference ID: %s", referenceId)
```

## Usage Examples

### 1. Basic Email with Attachments

```
client, err := maileroo.NewClient("your-api-key", 30)

if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

htmlContent := "<h1>Test Email</h1><p>This is a test email from the Maileroo Go SDK.</p>"
plainContent := "Test Email\n\nThis is a test email from the Maileroo Go SDK."

att1, err := maileroo.AttachmentFromContent("hello.txt", []byte("Hello, world!"), "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

att2, err := maileroo.AttachmentFromBase64Content("b64.txt", "SGVsbG8sIHdvcmxkIQ==", "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

att3, err := maileroo.AttachmentFromStream("stream.txt", strings.NewReader("Hello, world!"), "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

att4, err := maileroo.AttachmentFromFile("test_content.txt", "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

referenceId, err := client.SendBasicEmail(context.Background(), maileroo.BasicEmailData{
    From:        maileroo.NewEmail("YOUR_EMAIL_ADDRESS", "Your Name"),
    To:          []maileroo.EmailAddress{maileroo.NewEmail("RECIPIENT_EMAIL_ADDRESS", "Recipient Name")},
    Cc:          []maileroo.EmailAddress{maileroo.NewEmail("CC_EMAIL_ADDRESS", "CC Name")},
    Subject:     "Test Email from Maileroo Go SDK",
    HTML:        &htmlContent,
    Plain:       &plainContent,
    Attachments: []maileroo.Attachment{*att1, *att2, *att3, *att4},
    ReferenceID: maileroo.StrPtr(client.GetReferenceID()),
})

if err != nil {
    log.Fatalf("Failed to send email: %v", err)
}

log.Printf("Email sent with reference ID: %s", referenceId)
```

### 2. Template Email

```
client, err := maileroo.NewClient("your-api-key", 30)

if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

referenceId, err := client.SendTemplatedEmail(context.Background(), maileroo.TemplatedEmailData{
    From:        maileroo.NewEmail("YOUR_EMAIL_ADDRESS", "Your Name"),
    To:          []maileroo.EmailAddress{maileroo.NewEmail("RECIPIENT_EMAIL_ADDRESS", "Recipient Name")},
    Cc:          []maileroo.EmailAddress{maileroo.NewEmail("CC_EMAIL_ADDRESS", "CC Name")},
    Subject:    "Test Email from Maileroo Go SDK",
    TemplateID: 2549,
    TemplateData: map[string]any{
        "company": "Maileroo",
        "list": []any{
            map[string]any{
                "first_name": "John",
                "last_name":  "Doe",
            },
            map[string]any{
                "first_name": "Jane",
                "last_name":  "Smith",
            },
        },
    },
})

if err != nil {
    log.Fatalf("Failed to send email: %v", err)
}

log.Printf("Email sent with reference ID: %s", referenceId)
```

### 3. Bulk Email Sending (With Plain and HTML)

```
client, err := maileroo.NewClient("your-api-key", 30)

if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

att1, err := maileroo.AttachmentFromFile("test_content.txt", "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

referenceIds, err := client.SendBulkEmails(context.Background(), maileroo.BulkEmailData{
    Subject: "Test Email from Maileroo Go SDK",
    HTML:    maileroo.StrPtr("<html><body><h1>Hello, world!</h1></body></html>"),
    Plain:   maileroo.StrPtr("Hello, world!"),
    Messages: []maileroo.BulkMessage{
        {
            From: maileroo.NewEmail("SENDER_EMAIL_ADDRESS", "Sender Name"),
            To: []maileroo.EmailAddress{
                maileroo.NewEmail("RECIPIENT_EMAIL_ADDRESS", "Recipient Name"),
            },
            TemplateData: map[string]any{
                "name": "John Doe",
            },
            ReferenceID: maileroo.StrPtr(client.GetReferenceID()),
        },
        {
            From: maileroo.NewEmail("SENDER_EMAIL_ADDRESS", "Sender Name"),
            To: []maileroo.EmailAddress{
                maileroo.NewEmail("RECIPIENT_EMAIL_ADDRESS", "Recipient Name"),
            },
            TemplateData: map[string]any{
                "name": "Jane Doe",
            },
            ReferenceID: maileroo.StrPtr(client.GetReferenceID()),
        },
        // ... up to 500 messages
    },
    Tracking: maileroo.BoolPtr(true),
    Tags: maileroo.AssocMap{
        "tag1": "value1",
        "tag2": "value2",
    },
    Headers: maileroo.AssocMap{
        "X-Maileroo-Header": "value",
    },
    Attachments: []maileroo.Attachment{
        *att1,
    },
})

if err != nil {
    log.Fatalf("Failed to send emails: %v", err)
}

for _, referenceID := range referenceIds {
    log.Printf("Email sent with reference ID: %s", referenceID)
}
```

### 4. Bulk Email Sending (With Template ID)

```
client, err := maileroo.NewClient("your-api-key", 30)

if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

att1, err := maileroo.AttachmentFromFile("test_content.txt", "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

referenceIds, err := client.SendBulkEmails(context.Background(), maileroo.BulkEmailData{
    Subject:    "Test Email from Maileroo Go SDK",
    TemplateID: maileroo.IntPtr(2549),
    Messages: []maileroo.BulkMessage{
        {
            From: maileroo.NewEmail("SENDER_EMAIL_ADDRESS", "Sender Name"),
            To: []maileroo.EmailAddress{
                maileroo.NewEmail("RECIPIENT_EMAIL_ADDRESS", "Recipient Name"),
            },
            TemplateData: map[string]any{
                "name": "John Doe",
            },
            ReferenceID: maileroo.StrPtr(client.GetReferenceID()),
        },
        {
            From: maileroo.NewEmail("SENDER_EMAIL_ADDRESS", "Sender Name"),
            To: []maileroo.EmailAddress{
                maileroo.NewEmail("RECIPIENT_EMAIL_ADDRESS", "Recipient Name"),
            },
            TemplateData: map[string]any{
                "name": "Jane Doe",
            },
            ReferenceID: maileroo.StrPtr(client.GetReferenceID()),
        },
        // ... up to 500 messages
    },
    Tracking: maileroo.BoolPtr(false),
    Tags: maileroo.AssocMap{
        "tag1": "value1",
        "tag2": "value2",
    },
    Headers: maileroo.AssocMap{
        "X-Maileroo-Header": "value",
    },
    Attachments: []maileroo.Attachment{
        *att1,
    },
})

if err != nil {
    log.Fatalf("Failed to send emails: %v", err)
}

for _, referenceID := range referenceIds {
    log.Printf("Email sent with reference ID: %s", referenceID)
}
```

### 5. Working with Attachments

```
att1, err := maileroo.AttachmentFromContent("hello.txt", []byte("Hello, world!"), "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

att2, err := maileroo.AttachmentFromBase64Content("b64.txt", "SGVsbG8sIHdvcmxkIQ==", "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

att3, err := maileroo.AttachmentFromStream("stream.txt", strings.NewReader("Hello, world!"), "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

att4, err := maileroo.AttachmentFromFile("test_content.txt", "text/plain", false)

if err != nil {
    log.Fatalf("Failed to create attachment: %v", err)
}

attachments := []maileroo.Attachment{
    *att1,
    *att2,
    *att3,
    *att4,
}

for _, attachment := range attachments {
    log.Printf("Attachment Name: %s", attachment.FileName)
    log.Printf("Attachment Content Type: %s", attachment.ContentType)
    log.Printf("Attachment Content: %s", attachment.Content)
    log.Printf("Attachment Inline: %v", attachment.Inline)
}
```

### 6. Scheduling Emails

You can schedule emails for future delivery by adding a `ScheduledAt` field. It is available for both basic and template emails, but not for bulk emails.

```
func main() {

	client, err := maileroo.NewClient("your-api-key", 30)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	oneHourLater := time.Now().Add(time.Hour)

	htmlContent := "<h1>Test Email</h1><p>This is a test email from the Maileroo Go SDK.</p>"
	plainContent := "Test Email\n\nThis is a test email from the Maileroo Go SDK."

	referenceId, err := client.SendBasicEmail(context.Background(), maileroo.BasicEmailData{
		From: maileroo.NewEmail("YOUR_EMAIL_ADDRESS", "Your Name"),
		To: []maileroo.EmailAddress{
			maileroo.NewEmail("RECIPIENT_EMAIL_ADDRESS", "Recipient Name"),
		},
		Cc: []maileroo.EmailAddress{
			maileroo.NewEmail("CC_EMAIL_ADDRESS", "CC Name"),
		},
		Subject:     "Test Email from Maileroo Go SDK",
		HTML:        &htmlContent,
		Plain:       &plainContent,
		ScheduledAt: &oneHourLater,
	})

	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	log.Printf("Email sent with reference ID: %s", referenceId)

}
```

### 7. Managing Scheduled Emails

```
client, err := maileroo.NewClient("your-api-key", 30)

if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

response, err := client.GetScheduledEmails(context.Background(), 1, 100)

if err != nil {
    log.Fatalf("Failed to get scheduled emails: %v", err)
}

log.Printf("Page: %d", response.Page)
log.Printf("Per Page: %d", response.PerPage)
log.Printf("Total Pages: %d", response.TotalPages)

for _, item := range response.Items {
    log.Printf("%v", item)
}
```

### 8. Deleting Scheduled Email

```
client, err := maileroo.NewClient("your-api-key", 30)

if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

err = client.DeleteScheduledEmail(context.Background(), "your-reference-id")

if err != nil {
    log.Fatalf("Failed to delete scheduled email: %v", err)
} else {
    log.Println("Scheduled email deleted successfully")
}
```

## API Reference

### Client

#### Constructor

```
func NewClient(apiKey string, timeout int) (*Client, error)
```

#### Methods

- `SendBasicEmail(context.Context, BasicEmailData) (string, error)`
- `SendTemplateEmail(context.Context, TemplatedEmailData) (string, error)`
- `SendBulkEmails(context.Context, BulkEmailData) ([]string, error)`
- `GetScheduledEmails(context.Context, int, int) (*ScheduledEmailsResponse, error)`
- `DeleteScheduledEmail(context.Context, string) error`
- `GetReferenceID() string`

### EmailAddress

```
func NewEmail(address string, display_name string) EmailAddress
```

### Attachment

Static factory methods:

- `AttachmentFromContent(name string, content []byte, content_type string, inline bool) (*Attachment, error)`
- `AttachmentFromBase64Content(name string, content string, content_type string, inline bool) (*Attachment, error)`
- `AttachmentFromStream(name string, reader io.Reader, content_type string, inline bool) (*Attachment, error)`
- `AttachmentFromFile(name string, file_path string, content_type string, inline bool) (*Attachment, error)`

## Documentation

For detailed API documentation, including all available endpoints, parameters, and response formats, please refer to the [Maileroo API Documentation](https://maileroo.com/docs).

## License

This SDK is released under the MIT License.

## Support

Please visit our [support page](https://maileroo.com/contact-form) for any issues or questions regarding Maileroo. If you find any bugs or have feature requests, feel free to open an issue on our GitHub repository.