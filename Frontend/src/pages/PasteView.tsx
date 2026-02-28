import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import ContentLayout from "@cloudscape-design/components/content-layout";
import Header from "@cloudscape-design/components/header";
import Container from "@cloudscape-design/components/container";
import Box from "@cloudscape-design/components/box";
import Button from "@cloudscape-design/components/button";
import SpaceBetween from "@cloudscape-design/components/space-between";
import Spinner from "@cloudscape-design/components/spinner";
import Alert from "@cloudscape-design/components/alert";
import ColumnLayout from "@cloudscape-design/components/column-layout";
import { CodeView } from "@cloudscape-design/code-view";

interface Paste {
  id: string;
  title: string;
  content: string;
  language: string;
  created_at: string;
  expires_at: string | null;
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString();
}

export default function PasteView() {
  const { id } = useParams<{ id: string }>();
  const [paste, setPaste] = useState<Paste | null>(null);
  const [status, setStatus] = useState<"loading" | "ok" | "notfound" | "expired" | "error">(
    "loading",
  );
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!id) return;
    fetch(`/api/paste/${id}`)
      .then((res) => {
        if (res.status === 404) {
          setStatus("notfound");
          return null;
        }
        if (res.status === 410) {
          setStatus("expired");
          return null;
        }
        if (!res.ok) {
          setStatus("error");
          return null;
        }
        return res.json();
      })
      .then((data) => {
        if (data) {
          setPaste(data);
          setStatus("ok");
        }
      })
      .catch(() => setStatus("error"));
  }, [id]);

  function handleCopy() {
    if (!paste) return;
    navigator.clipboard.writeText(paste.content).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }

  if (status === "loading") {
    return (
      <ContentLayout header={<Header variant="h1">Loading…</Header>}>
        <Container>
          <Spinner size="large" />
        </Container>
      </ContentLayout>
    );
  }

  if (status === "notfound") {
    return (
      <ContentLayout header={<Header variant="h1">Not Found</Header>}>
        <Container>
          <SpaceBetween size="m">
            <Alert type="error">This paste does not exist.</Alert>
            <Link to="/paste">Create a new paste</Link>
          </SpaceBetween>
        </Container>
      </ContentLayout>
    );
  }

  if (status === "expired") {
    return (
      <ContentLayout header={<Header variant="h1">Expired</Header>}>
        <Container>
          <SpaceBetween size="m">
            <Alert type="warning">This paste has expired and is no longer available.</Alert>
            <Link to="/paste">Create a new paste</Link>
          </SpaceBetween>
        </Container>
      </ContentLayout>
    );
  }

  if (status === "error" || !paste) {
    return (
      <ContentLayout header={<Header variant="h1">Error</Header>}>
        <Container>
          <Alert type="error">Failed to load paste. Please try again.</Alert>
        </Container>
      </ContentLayout>
    );
  }

  return (
    <ContentLayout
      header={
        <Header
          variant="h1"
          actions={
            <SpaceBetween direction="horizontal" size="xs">
              <Button onClick={handleCopy} iconName={copied ? "status-positive" : "copy"}>
                {copied ? "Copied!" : "Copy"}
              </Button>
              <Button href={`/api/paste/${id}/raw`} target="_blank" iconName="external">
                Raw
              </Button>
            </SpaceBetween>
          }
        >
          {paste.title || "Untitled"}
        </Header>
      }
    >
      <SpaceBetween size="m">
        <Container>
          <ColumnLayout columns={3} variant="text-grid">
            <div>
              <Box variant="awsui-key-label">Language</Box>
              <Box>{paste.language || "Plain Text"}</Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Created</Box>
              <Box>{formatDate(paste.created_at)}</Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Expires</Box>
              <Box>{paste.expires_at ? formatDate(paste.expires_at) : "Never"}</Box>
            </div>
          </ColumnLayout>
        </Container>
        <Container>
          <CodeView content={paste.content} lineNumbers wrapLines />
        </Container>
      </SpaceBetween>
    </ContentLayout>
  );
}
