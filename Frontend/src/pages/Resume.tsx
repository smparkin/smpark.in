import ContentLayout from "@cloudscape-design/components/content-layout";
import Header from "@cloudscape-design/components/header";
import Container from "@cloudscape-design/components/container";
import SpaceBetween from "@cloudscape-design/components/space-between";
import Box from "@cloudscape-design/components/box";
import ColumnLayout from "@cloudscape-design/components/column-layout";

export default function Resume() {
  return (
    <ContentLayout header={<Header variant="h1">Resume</Header>}>
      <SpaceBetween size="l">
        <Container header={<Header variant="h2">Employment</Header>}>
          <SpaceBetween size="l">
            <SpaceBetween size="xxs">
              <Header variant="h3">Systems Development Engineer, Amazon Security</Header>
              <Box variant="small" color="text-body-secondary">
                San Luis Obispo, CA &middot; April 2020 &ndash; Present (Promoted to full-time July
                2022)
              </Box>
              <Box variant="p">
                <ul>
                  <li>
                    Create and deploy serverless applications to collect and enrich data for
                    analysis.
                  </li>
                  <li>
                    Add monitoring and alarming to existing services to allow for easier diagnosing
                    of potential issues.
                  </li>
                  <li>
                    Work with customers to understand requirements and deliver results quickly.
                  </li>
                </ul>
              </Box>
            </SpaceBetween>

            <SpaceBetween size="xxs">
              <Header variant="h3">Student Assistant, On-Site Support Cal Poly ITS</Header>
              <Box variant="small" color="text-body-secondary">
                San Luis Obispo, CA &middot; May 2019 &ndash; February 2020
              </Box>
              <Box variant="p">
                <ul>
                  <li>
                    Responsible for support of computer services and equipment across multiple
                    departments.
                  </li>
                  <li>Troubleshooting, detecting, and solving of technical problems.</li>
                  <li>
                    Managing macOS and Windows based computers using Active Directory, SCCM, and
                    Jamf.
                  </li>
                </ul>
              </Box>
            </SpaceBetween>
          </SpaceBetween>
        </Container>

        <Container header={<Header variant="h2">Education</Header>}>
          <SpaceBetween size="l">
            <SpaceBetween size="xxs">
              <Header variant="h3">California Polytechnic State University</Header>
              <Box variant="small" color="text-body-secondary">
                2017 &ndash; 2022
              </Box>
              <Box variant="p">
                <strong>Software Engineering</strong>
                <br />
                Relevant Coursework: Introduction to Computing, Fundamentals of Computer Science,
                Data Structures, Project-Based Object-Oriented Programming &amp; Design,
                Introduction to Computer Organization, Systems Programming, and Introduction to
                Operating Systems.
              </Box>
            </SpaceBetween>

            <SpaceBetween size="xxs">
              <Header variant="h3">Cal Poly Security Education Club &mdash; President</Header>
              <Box variant="small" color="text-body-secondary">
                Spring 2020 &ndash; Spring 2021
              </Box>
              <Box variant="p">
                <ul>
                  <li>
                    Responsible for planning and logistics of events including iFixit Triathlon and
                    Security Career Fair.
                  </li>
                  <li>
                    Manage a team of student officers to introduce students to security topics at
                    all skill levels.
                  </li>
                  <li>Coordinate with companies for presentations and special events.</li>
                  <li>
                    Presented technical talks such as "iOS Security and Jailbreaking", "machswap, a
                    vulnerability in XNU IPC", "Intro to SSH", "Nintendo Switch Security", "macOS
                    Security", and "Mac File Systems and APFS Security".
                  </li>
                </ul>
              </Box>
            </SpaceBetween>
          </SpaceBetween>
        </Container>

        <Container header={<Header variant="h2">Skills</Header>}>
          <ColumnLayout columns={3} variant="text-grid">
            <SpaceBetween size="xxs">
              <Box variant="h3">Languages &amp; Tools</Box>
              <Box variant="p">
                Python, TypeScript, React, JavaScript, C, Swift, git, zsh, Burp Suite
              </Box>
            </SpaceBetween>
            <SpaceBetween size="xxs">
              <Box variant="h3">AWS</Box>
              <Box variant="p">Lambda, S3, DynamoDB, CloudWatch, KMS, SSM, SES, SQS, SNS, CDK</Box>
            </SpaceBetween>
            <SpaceBetween size="xxs">
              <Box variant="h3">Operating Systems</Box>
              <Box variant="p">macOS, Windows, Linux</Box>
            </SpaceBetween>
          </ColumnLayout>
        </Container>

        <Container header={<Header variant="h2">Projects</Header>}>
          <ul>
            <li>
              Developed a watchOS app to show the status of the White Hat lab on an Apple Watch.
            </li>
            <li>Created a Python script to control Spotify's API from the command line.</li>
            <li>
              Upgraded White Hat's network infrastructure to take advantage of the 1Gbps connection
              to the internet.
            </li>
          </ul>
        </Container>
      </SpaceBetween>
    </ContentLayout>
  );
}
