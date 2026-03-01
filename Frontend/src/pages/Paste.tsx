import { useState } from "react";
import { useNavigate } from "react-router-dom";
import ContentLayout from "@cloudscape-design/components/content-layout";
import Header from "@cloudscape-design/components/header";
import Container from "@cloudscape-design/components/container";
import FormField from "@cloudscape-design/components/form-field";
import Input from "@cloudscape-design/components/input";
import Textarea from "@cloudscape-design/components/textarea";
import Select from "@cloudscape-design/components/select";
import Slider from "@cloudscape-design/components/slider";
import Button from "@cloudscape-design/components/button";
import Alert from "@cloudscape-design/components/alert";
import SpaceBetween from "@cloudscape-design/components/space-between";

const LANGUAGE_OPTIONS = [
  { label: "Plain Text", value: "plaintext" },
  { label: "Bash", value: "bash" },
  { label: "C", value: "c" },
  { label: "C++", value: "cpp" },
  { label: "CSS", value: "css" },
  { label: "Dockerfile", value: "dockerfile" },
  { label: "Go", value: "go" },
  { label: "HTML", value: "html" },
  { label: "Java", value: "java" },
  { label: "JavaScript", value: "javascript" },
  { label: "JSON", value: "json" },
  { label: "Kotlin", value: "kotlin" },
  { label: "Markdown", value: "markdown" },
  { label: "Python", value: "python" },
  { label: "Ruby", value: "ruby" },
  { label: "Rust", value: "rust" },
  { label: "SQL", value: "sql" },
  { label: "Swift", value: "swift" },
  { label: "TypeScript", value: "typescript" },
  { label: "YAML", value: "yaml" },
];

export default function Paste() {
  const navigate = useNavigate();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [language, setLanguage] = useState(LANGUAGE_OPTIONS[0]);
  const [expiry, setExpiry] = useState(5);
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit() {
    if (!content.trim()) {
      setError("Content is required.");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/paste", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: title.trim(),
          content,
          language: language.value,
          expiry,
          password: password || undefined,
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error ?? "Failed to create paste.");
        return;
      }
      navigate(`/paste/${data.id}`, { state: { password: password || undefined } });
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <ContentLayout header={<Header variant="h1">New Paste</Header>}>
      <Container>
        <SpaceBetween size="m">
          {error && (
            <Alert type="error" onDismiss={() => setError(null)} dismissible>
              {error}
            </Alert>
          )}
          <FormField label="Title" description="Optional title for your paste.">
            <Input
              value={title}
              onChange={(e) => setTitle(e.detail.value)}
              placeholder="Untitled"
            />
          </FormField>
          <FormField label="Content" constraintText="Required. Maximum 1 MB.">
            <Textarea
              value={content}
              onChange={(e) => setContent(e.detail.value)}
              placeholder="Paste your content here..."
              rows={20}
              spellcheck={false}
            />
          </FormField>
          <FormField label="Language">
            <Select
              selectedOption={language}
              onChange={(e) =>
                setLanguage(e.detail.selectedOption as { label: string; value: string })
              }
              options={LANGUAGE_OPTIONS}
            />
          </FormField>
          <FormField label="Password" description="Optional. Leave blank for a public paste.">
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.detail.value)}
              placeholder="Leave blank for no password"
            />
          </FormField>
          <FormField label="Expires" description={`${expiry} minute${expiry === 1 ? "" : "s"}`}>
            <Slider
              value={expiry}
              onChange={(e) => setExpiry(e.detail.value)}
              min={1}
              max={10}
              tickMarks
              valueFormatter={(v) => `${v}m`}
            />
          </FormField>
          <Button variant="primary" onClick={handleSubmit} loading={loading}>
            Create Paste
          </Button>
        </SpaceBetween>
      </Container>
    </ContentLayout>
  );
}
