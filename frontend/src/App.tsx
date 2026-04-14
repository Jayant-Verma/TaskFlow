import {
  BrowserRouter,
  Routes,
  Route,
  Navigate,
  useParams,
} from 'react-router-dom';
import { Box, Container, Heading } from '@chakra-ui/react';
import { AuthPage } from '@/features/auth/AuthPage';
import { KanbanBoard } from '@/features/tasks/KanbanBoard';
import { Navbar } from '@/components/Navbar';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { ProjectsList } from './features/projects/ProjectsList';

const ProjectDetailView = () => {
  const { id } = useParams();
  return (
    <>
      <Heading size="lg" mb={2}>
        Project Board
      </Heading>
      <KanbanBoard projectId={id!} />
    </>
  );
};

export default function App() {
  const isAuth = !!localStorage.getItem('taskflow_token');

  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/login"
          element={isAuth ? <Navigate to="/projects" /> : <AuthPage />}
        />

        {/* Protected Dashboard Layout */}
        <Route
          path="/*"
          element={
            <ProtectedRoute>
              <Box minH="100vh" bg="white" _dark={{ bg: 'gray.900' }}>
                <Navbar />
                <Container maxW="container.xl" py={8}>
                  <Routes>
                    <Route path="/projects" element={<ProjectsList />} />
                    <Route
                      path="/projects/:id"
                      element={<ProjectDetailView />}
                    />
                    <Route
                      path="*"
                      element={<Navigate to="/projects" replace />}
                    />
                  </Routes>
                </Container>
              </Box>
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}
