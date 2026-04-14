import {
  Box,
  SimpleGrid,
  Heading,
  Button,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  FormControl,
  FormLabel,
  Input,
  FormErrorMessage,
  VStack,
  useToast,
  Text,
  HStack,
  Spinner,
  Center,
} from '@chakra-ui/react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/api/client';
import { Project } from '@/types';
import { useNavigate } from 'react-router-dom';
import { Plus, FolderKanban } from 'lucide-react';

const projectSchema = z.object({
  name: z.string().min(1, 'Project name is required'),
  description: z.string().optional(),
});

type ProjectFormValues = z.infer<typeof projectSchema>;

export const ProjectsList = () => {
  const { isOpen, onOpen, onClose } = useDisclosure();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const toast = useToast();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<ProjectFormValues>({
    resolver: zodResolver(projectSchema),
  });

  // Fetch Projects
  const {
    data: projects = [],
    isLoading,
    isError,
  } = useQuery<Project[]>({
    queryKey: ['projects'],
    queryFn: async () => (await api.get('/projects')).data,
  });

  // Create Project
  const { mutate: createProject } = useMutation({
    mutationFn: async (data: ProjectFormValues) =>
      await api.post('/projects', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      toast({ title: 'Project created.', status: 'success' });
      onClose();
      reset();
    },
    onError: () =>
      toast({ title: 'Failed to create project', status: 'error' }),
  });

  if (isLoading)
    return (
      <Center h="50vh">
        <Spinner size="xl" color="blue.500" />
      </Center>
    );
  if (isError)
    return (
      <Center h="50vh">
        <Text color="red.500">Failed to load projects.</Text>
      </Center>
    );

  return (
    <Box>
      <HStack justify="space-between" mb={8}>
        <Heading size="lg" display="flex" alignItems="center" gap={3}>
          <FolderKanban /> My Projects
        </Heading>
        <Button
          leftIcon={<Plus size={18} />}
          colorScheme="blue"
          onClick={onOpen}
        >
          New Project
        </Button>
      </HStack>

      {projects.length === 0 ? (
        <Center
          p={10}
          border="2px dashed"
          borderColor="gray.300"
          _dark={{ borderColor: 'gray.600' }}
          borderRadius="xl"
        >
          <VStack spacing={3}>
            <Text color="gray.500">No projects found.</Text>
            <Button variant="outline" size="sm" onClick={onOpen}>
              Create your first project
            </Button>
          </VStack>
        </Center>
      ) : (
        <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={6}>
          {projects.map((project) => (
            <Box
              key={project.id}
              p={6}
              bg="white"
              _dark={{ bg: 'gray.800', borderColor: 'gray.700' }}
              borderWidth={1}
              borderRadius="xl"
              boxShadow="sm"
              cursor="pointer"
              transition="all 0.2s"
              _hover={{
                transform: 'translateY(-2px)',
                boxShadow: 'md',
                borderColor: 'blue.400',
              }}
              onClick={() => navigate(`/projects/${project.id}`)}
            >
              <Heading size="md" mb={2}>
                {project.name}
              </Heading>
              <Text
                color="gray.600"
                _dark={{ color: 'gray.400' }}
                noOfLines={2}
              >
                {project.description || 'No description provided.'}
              </Text>
            </Box>
          ))}
        </SimpleGrid>
      )}

      {/* Create Project Modal */}
      <Modal isOpen={isOpen} onClose={onClose} isCentered>
        <ModalOverlay backdropFilter="blur(4px)" />
        <ModalContent>
          <ModalHeader>Create New Project</ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <VStack
              as="form"
              id="project-form"
              onSubmit={handleSubmit((data) => createProject(data))}
              spacing={4}
            >
              <FormControl isInvalid={!!errors.name}>
                <FormLabel>Project Name</FormLabel>
                <Input
                  {...register('name')}
                  placeholder="e.g. Website Redesign"
                />
                <FormErrorMessage>{errors.name?.message}</FormErrorMessage>
              </FormControl>
              <FormControl isInvalid={!!errors.description}>
                <FormLabel>Description</FormLabel>
                <Input
                  {...register('description')}
                  placeholder="Brief details..."
                />
                <FormErrorMessage>
                  {errors.description?.message}
                </FormErrorMessage>
              </FormControl>
            </VStack>
          </ModalBody>
          <Box px={6} pb={6} display="flex" justifyContent="flex-end" gap={3}>
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              form="project-form"
              colorScheme="blue"
              isLoading={isSubmitting}
            >
              Create
            </Button>
          </Box>
        </ModalContent>
      </Modal>
    </Box>
  );
};
