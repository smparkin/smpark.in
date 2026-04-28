import ContentLayout from "@cloudscape-design/components/content-layout";
import Header from "@cloudscape-design/components/header";
import Box from "@cloudscape-design/components/box";
import Link from "@cloudscape-design/components/link";

export default function NotFound() {
  return (
    <ContentLayout header={<Header variant="h1">404</Header>}>
      <Box variant="p">
        The page you're looking for doesn't exist. <Link href="/">Go home</Link>
      </Box>
    </ContentLayout>
  );
}
