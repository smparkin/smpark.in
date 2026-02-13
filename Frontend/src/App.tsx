import { BrowserRouter, Routes, Route, useNavigate, useLocation } from "react-router-dom";
import AppLayout from "@cloudscape-design/components/app-layout";
import SideNavigation from "@cloudscape-design/components/side-navigation";
import "@cloudscape-design/global-styles/index.css";
import "./App.css";
import Home from "./pages/Home";
import Resume from "./pages/Resume";
import NotFound from "./pages/NotFound";

function AppContent() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <AppLayout
      toolsHide
      navigation={
        <SideNavigation
          header={{ text: "Stephen Parkinson", href: "/" }}
          activeHref={location.pathname}
          onFollow={(event) => {
            event.preventDefault();
            navigate(event.detail.href);
          }}
          items={[
            { type: "link", text: "Home", href: "/" },
            { type: "link", text: "Resume", href: "/resume" },
          ]}
        />
      }
      content={
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/resume" element={<Resume />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      }
    />
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AppContent />
    </BrowserRouter>
  );
}
