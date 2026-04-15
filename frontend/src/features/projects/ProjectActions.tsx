import { useRef } from 'react';
import {
  Box,
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
  HStack,
  useToast,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
} from '@chakra-ui/react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/api/client';
import { useNavigate } from 'react-router-dom';
import { Settings, Trash2 } from 'lucide-react';
import { Project } from '@/types';

const projectSchema = z.object({
  name: z.string().min(1, 'Project name is required'),
  description: z.string().optional(),
});

export const ProjectActions = ({ projectId }: { projectId: string }) => {
  const queryClient = useQueryClient();
  const toast = useToast();
  const navigate = useNavigate();

  const {
    isOpen: isEditOpen,
    onOpen: onEditOpen,
    onClose: onEditClose,
  } = useDisclosure();
  const {
    isOpen: isAlertOpen,
    onOpen: onAlertOpen,
    onClose: onAlertClose,
  } = useDisclosure();
  const cancelRef = useRef<HTMLButtonElement>(null);

  const { data: project, isFetching } = useQuery<Project>({
    queryKey: ['project', projectId],
    queryFn: async () => (await api.get(`/projects/${projectId}`)).data.project,
    enabled: isEditOpen,
  });

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<z.infer<typeof projectSchema>>({
    resolver: zodResolver(projectSchema),
    values: {
      name: project?.name || '',
      description: project?.description || '',
    },
  });

  // PATCH: Update Project
  const { mutate: updateProject } = useMutation({
    mutationFn: async (data: z.infer<typeof projectSchema>) =>
      await api.patch(`/projects/${projectId}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      toast({ title: 'Project updated.', status: 'success' });
      onEditClose();
    },
    onError: () =>
      toast({ title: 'Failed to update project', status: 'error' }),
  });

  // DELETE: Remove Project
  const { mutate: deleteProject, isPending: isDeleting } = useMutation({
    mutationFn: async () => await api.delete(`/projects/${projectId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      toast({ title: 'Project deleted.', status: 'info' });
      navigate('/projects');
    },
    onError: () => {
      toast({ title: 'Failed to delete project', status: 'error' });
      onAlertClose();
    },
  });

  return (
    <>
      <HStack spacing={3}>
        <Button
          size="sm"
          variant="outline"
          leftIcon={<Settings size={16} />}
          onClick={onEditOpen}
        >
          Edit Project
        </Button>
        <Button
          size="sm"
          colorScheme="red"
          variant="ghost"
          leftIcon={<Trash2 size={16} />}
          onClick={onAlertOpen}
        >
          Delete
        </Button>
      </HStack>

      <Modal isOpen={isEditOpen} onClose={onEditClose} isCentered>
        <ModalOverlay backdropFilter="blur(4px)" />
        <ModalContent>
          <ModalHeader>Edit Project Details</ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <VStack
              as="form"
              id="edit-project-form"
              onSubmit={handleSubmit((data) => updateProject(data))}
              spacing={4}
            >
              <FormControl isInvalid={!!errors.name}>
                <FormLabel>Project Name</FormLabel>
                <Input {...register('name')} />
                <FormErrorMessage>{errors.name?.message}</FormErrorMessage>
              </FormControl>
              <FormControl isInvalid={!!errors.description}>
                <FormLabel>Description</FormLabel>
                <Input {...register('description')} />
                <FormErrorMessage>
                  {errors.description?.message}
                </FormErrorMessage>
              </FormControl>
            </VStack>
          </ModalBody>
          <Box px={6} pb={6} display="flex" justifyContent="flex-end" gap={3}>
            <Button variant="ghost" onClick={onEditClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              form="edit-project-form"
              colorScheme="blue"
              isLoading={isSubmitting || isFetching}
            >
              Save Changes
            </Button>
          </Box>
        </ModalContent>
      </Modal>

      <AlertDialog
        isOpen={isAlertOpen}
        leastDestructiveRef={cancelRef}
        onClose={onAlertClose}
        isCentered
      >
        <AlertDialogOverlay backdropFilter="blur(4px)">
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              Delete Project
            </AlertDialogHeader>
            <AlertDialogBody>
              Are you sure? This will permanently delete{' '}
              <strong>{project?.name}</strong> and all of its tasks. This action
              cannot be undone.
            </AlertDialogBody>
            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onAlertClose}>
                Cancel
              </Button>
              <Button
                colorScheme="red"
                onClick={() => deleteProject()}
                isLoading={isDeleting}
                ml={3}
              >
                Yes, Delete Project
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </>
  );
};
