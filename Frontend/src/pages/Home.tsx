import ContentLayout from "@cloudscape-design/components/content-layout";
import Header from "@cloudscape-design/components/header";
import Container from "@cloudscape-design/components/container";
import SpaceBetween from "@cloudscape-design/components/space-between";
import Link from "@cloudscape-design/components/link";
import Box from "@cloudscape-design/components/box";

export default function Home() {
  return (
    <ContentLayout
      header={
        <div className="page-header">
          <img src="/images/me.jpg" alt="Stephen Parkinson" className="profile-image" />
          <Header variant="h1" description="Systems Development Engineer">
            Stephen Parkinson
          </Header>
        </div>
      }
    >
      <SpaceBetween size="l">
        <Container header={<Header variant="h2">About Me</Header>}>
          <Box variant="p">I'm doing a lot of stuff with computers right now.</Box>
        </Container>

        <Container>
          <img src="/images/trees.jpg" alt="Trees" className="hero-image" />
        </Container>

        <Container header={<Header variant="h2">Contact</Header>}>
          <SpaceBetween size="s" direction="horizontal">
            <Link href="https://www.linkedin.com/in/stephen-parkinson" external>
              LinkedIn
            </Link>
            <Link href="https://github.com/smparkin" external>
              GitHub
            </Link>
          </SpaceBetween>
        </Container>
      </SpaceBetween>
    </ContentLayout>
  );
}
