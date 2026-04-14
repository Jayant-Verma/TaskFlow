import { useState } from 'react';
import {
  DragDropContext,
  Droppable,
  Draggable,
  DropResult,
} from '@hello-pangea/dnd';
import {
  Box,
  SimpleGrid,
  Text,
  Badge,
  VStack,
  HStack,
  Spinner,
  Center,
  useToast,
  Button,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalCloseButton,
  ModalBody,
  FormControl,
  FormLabel,
  Input,
  FormErrorMessage,
  Select,
  Textarea,
  Flex,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
} from '@chakra-ui/react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/api/client';
import { Task, User as UserType } from '@/types';
import { Plus, Calendar, User } from 'lucide-react';
import { ProjectStatsHeader } from './ProjectStatsHeader';
import { Link as RouterLink } from 'react-router-dom';

const COLUMNS = [
  { id: 'todo', label: 'To Do', color: 'gray' },
  { id: 'in_progress', label: 'In Progress', color: 'blue' },
  { id: 'done', label: 'Done', color: 'green' },
] as const;

const taskSchema = z.object({
  title: z.string().min(1, 'Task title is required'),
  description: z.string().optional(),
  priority: z.enum(['low', 'medium', 'high']),
  assignee_id: z.string().optional(),
  due_date: z.string().optional(),
});

type TaskFormValues = z.infer<typeof taskSchema>;

export const KanbanBoard = ({ projectId }: { projectId: string }) => {
  const queryClient = useQueryClient();
  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();

  // State for Filters & Editing
  const [filterAssignee, setFilterAssignee] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [editingTask, setEditingTask] = useState<Task | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<TaskFormValues>({
    resolver: zodResolver(taskSchema),
    defaultValues: { priority: 'medium' },
  });

  // --- Queries & Mutations ---
  const {
    data: tasks = [],
    isLoading,
    isError,
  } = useQuery<Task[]>({
    queryKey: ['tasks', projectId],
    queryFn: async () => (await api.get(`/projects/${projectId}/tasks`)).data,
  });

  const { data: members = [] } = useQuery<UserType[]>({
    queryKey: ['users'],
    queryFn: async () => (await api.get(`/users`)).data,
  });

  const { mutate: saveTask } = useMutation({
    mutationFn: async (data: TaskFormValues) => {
      const cleanPayload: Partial<TaskFormValues> = {};

      Object.entries(data).forEach(([key, value]) => {
        if (value !== '' && value !== null && value !== undefined) {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          cleanPayload[key as keyof TaskFormValues] = value as any;
        }
      });
      if (editingTask)
        return await api.patch(`/tasks/${editingTask.id}`, cleanPayload);
      return await api.post(`/projects/${projectId}/tasks`, cleanPayload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks', projectId] });
      toast({
        title: `Task ${editingTask ? 'updated' : 'created'}.`,
        status: 'success',
      });
      handleCloseModal();
    },
    onError: () => toast({ title: 'Failed to save task', status: 'error' }),
  });

  const { mutate: updateStatus } = useMutation({
    mutationFn: async ({ id, status }: { id: string; status: string }) =>
      await api.patch(`/tasks/${id}`, { status }),
    onMutate: async ({ id, status }) => {
      await queryClient.cancelQueries({ queryKey: ['tasks', projectId] });
      const prevTasks = queryClient.getQueryData<Task[]>(['tasks', projectId]);
      queryClient.setQueryData<Task[]>(['tasks', projectId], (old) =>
        old?.map((t) =>
          t.id === id
            ? {
                ...t,
                status: status as 'todo' | 'in_progress' | 'done',
              }
            : t
        )
      );
      return { prevTasks };
    },
    onError: (_err, _vars, context) => {
      queryClient.setQueryData(['tasks', projectId], context?.prevTasks);
      toast({ title: 'Failed to move task', status: 'error' });
    },
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: ['tasks', projectId] }),
  });

  // --- Handlers ---
  const handleOpenCreate = () => {
    setEditingTask(null);
    reset({
      title: '',
      description: '',
      priority: 'medium',
      assignee_id: '',
      due_date: '',
    });
    onOpen();
  };

  const handleOpenEdit = (task: Task) => {
    setEditingTask(task);
    reset({
      title: task.title,
      description: task.description,
      priority: task.priority,
      assignee_id: task.assignee_id || '',
      due_date: task.due_date ? task.due_date.split('T')[0] : '', // Format for date input
    });
    onOpen();
  };

  const handleCloseModal = () => {
    onClose();
    setEditingTask(null);
    reset();
  };

  const onDragEnd = (result: DropResult) => {
    if (
      !result.destination ||
      result.destination.droppableId === result.source.droppableId
    )
      return;
    updateStatus({
      id: result.draggableId,
      status: result.destination.droppableId,
    });
  };

  if (isLoading)
    return (
      <Center h="50vh">
        <Spinner size="xl" color="blue.500" />
      </Center>
    );
  if (isError)
    return (
      <Center h="50vh">
        <Text color="red.500">Failed to load tasks.</Text>
      </Center>
    );

  // Apply Filters
  const filteredTasks = tasks.filter((task) => {
    if (filterAssignee && task.assignee_id !== filterAssignee) return false;
    return true;
  });

  const visibleColumns = filterStatus
    ? COLUMNS.filter((c) => c.id === filterStatus)
    : COLUMNS;

  return (
    <Box>
      <Box mb={6}>
        <Breadcrumb fontSize="sm" color="gray.500">
          <BreadcrumbItem>
            <BreadcrumbLink as={RouterLink} to="/projects">
              Projects
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbItem isCurrentPage>
            <BreadcrumbLink fontWeight="bold">Board</BreadcrumbLink>
          </BreadcrumbItem>
        </Breadcrumb>
      </Box>
      <ProjectStatsHeader projectId={projectId} />
      {/* FILTER BAR & ACTION */}
      <Flex
        justify="space-between"
        align="center"
        mb={6}
        flexWrap="wrap"
        gap={4}
      >
        <HStack spacing={4} w={{ base: 'full', md: 'auto' }}>
          <Select
            size="sm"
            placeholder="All Assignees"
            value={filterAssignee}
            onChange={(e) => setFilterAssignee(e.target.value)}
          >
            {members.map((u) => (
              <option key={u.id} value={u.id}>
                {u.name}
              </option>
            ))}
          </Select>
          <Select
            size="sm"
            placeholder="All Statuses"
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            bg="white"
            _dark={{ bg: 'gray.800' }}
          >
            {COLUMNS.map((c) => (
              <option key={c.id} value={c.id}>
                {c.label}
              </option>
            ))}
          </Select>
        </HStack>
        <Button
          leftIcon={<Plus size={18} />}
          colorScheme="blue"
          size="sm"
          w={{ base: 'full', md: 'auto' }}
          onClick={handleOpenCreate}
        >
          Add Task
        </Button>
      </Flex>

      {/* KANBAN BOARD */}
      <DragDropContext onDragEnd={onDragEnd}>
        <SimpleGrid
          columns={{ base: 1, md: visibleColumns.length }}
          spacing={6}
        >
          {visibleColumns.map((col) => {
            const columnTasks = filteredTasks.filter(
              (t) => t.status === col.id
            );

            return (
              <Box
                key={col.id}
                bg="gray.50"
                _dark={{ bg: 'whiteAlpha.50' }}
                p={4}
                borderRadius="xl"
                minH="65vh"
              >
                <HStack justify="space-between" mb={4}>
                  <Text
                    fontWeight="bold"
                    textTransform="uppercase"
                    fontSize="sm"
                  >
                    {col.label}
                  </Text>
                  <Badge colorScheme={col.color} borderRadius="full" px={2}>
                    {columnTasks.length}
                  </Badge>
                </HStack>

                <Droppable droppableId={col.id}>
                  {(provided, snapshot) => (
                    <VStack
                      {...provided.droppableProps}
                      ref={provided.innerRef}
                      spacing={3}
                      align="stretch"
                      h="full"
                      p={1}
                      borderRadius="md"
                      bg={snapshot.isDraggingOver ? 'blue.50' : 'transparent'}
                      _dark={{
                        bg: snapshot.isDraggingOver
                          ? 'whiteAlpha.100'
                          : 'transparent',
                      }}
                      transition="background 0.2s"
                    >
                      {columnTasks.map((task, index) => (
                        <Draggable
                          key={task.id}
                          draggableId={task.id}
                          index={index}
                        >
                          {(provided, snap) => (
                            <Box
                              ref={provided.innerRef}
                              {...provided.draggableProps}
                              {...provided.dragHandleProps}
                              bg="white"
                              p={4}
                              borderRadius="md"
                              shadow={snap.isDragging ? 'lg' : 'sm'}
                              border="1px"
                              borderColor={
                                snap.isDragging ? 'blue.400' : 'gray.200'
                              }
                              _dark={{
                                bg: 'gray.700',
                                borderColor: 'gray.600',
                              }}
                              onClick={() => handleOpenEdit(task)} // EDIT TRIGGER
                              cursor="pointer"
                              _hover={{
                                borderColor: 'blue.300',
                              }}
                              transition="border 0.2s"
                            >
                              <Text fontWeight="semibold" mb={1}>
                                {task.title}
                              </Text>
                              <HStack justify="space-between" mt={3}>
                                <Badge
                                  variant="subtle"
                                  colorScheme={
                                    task.priority === 'high'
                                      ? 'red'
                                      : task.priority === 'medium'
                                        ? 'orange'
                                        : 'gray'
                                  }
                                >
                                  {task.priority}
                                </Badge>
                                {task.assignee_id && (
                                  <HStack spacing={1}>
                                    <User size={12} />
                                    <Text fontSize="xs" color="gray.500">
                                      {members.find(
                                        (m) => m.id === task.assignee_id
                                      )?.name || 'Assigned'}
                                    </Text>
                                  </HStack>
                                )}
                              </HStack>
                            </Box>
                          )}
                        </Draggable>
                      ))}
                      {provided.placeholder}
                      {columnTasks.length === 0 && (
                        <Center
                          p={4}
                          border="2px dashed"
                          borderColor="gray.300"
                          _dark={{
                            borderColor: 'gray.600',
                          }}
                          borderRadius="md"
                        >
                          <Text fontSize="sm" color="gray.400">
                            Drop tasks here
                          </Text>
                        </Center>
                      )}
                    </VStack>
                  )}
                </Droppable>
              </Box>
            );
          })}
        </SimpleGrid>
      </DragDropContext>

      {/* CREATE / EDIT MODAL */}
      <Modal isOpen={isOpen} onClose={handleCloseModal} isCentered size="lg">
        <ModalOverlay backdropFilter="blur(4px)" />
        <ModalContent>
          <ModalHeader>
            {editingTask ? 'Edit Task' : 'Create New Task'}
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <VStack
              as="form"
              id="task-form"
              onSubmit={handleSubmit((data) => saveTask(data))}
              spacing={4}
            >
              <FormControl isInvalid={!!errors.title}>
                <FormLabel>Task Title</FormLabel>
                <Input
                  {...register('title')}
                  placeholder="e.g. Setup database schema"
                  autoFocus
                />
                <FormErrorMessage>{errors.title?.message}</FormErrorMessage>
              </FormControl>

              <FormControl isInvalid={!!errors.description}>
                <FormLabel>Description</FormLabel>
                <Textarea
                  {...register('description')}
                  placeholder="Add details..."
                  resize="none"
                />
                <FormErrorMessage>
                  {errors.description?.message}
                </FormErrorMessage>
              </FormControl>

              <SimpleGrid columns={2} spacing={4} w="full">
                <FormControl isInvalid={!!errors.priority}>
                  <FormLabel>Priority</FormLabel>
                  <Select {...register('priority')}>
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                  </Select>
                  <FormErrorMessage>
                    {errors.priority?.message}
                  </FormErrorMessage>
                </FormControl>

                <FormControl isInvalid={!!errors.due_date}>
                  <FormLabel display="flex" alignItems="center" gap={2}>
                    <Calendar size={14} /> Due Date
                  </FormLabel>
                  <Input
                    type="date"
                    {...register('due_date')}
                    css={{
                      '::-webkit-calendar-picker-indicator': {
                        cursor: 'pointer',
                      },
                    }}
                  />
                  <FormErrorMessage>
                    {errors.due_date?.message}
                  </FormErrorMessage>
                </FormControl>
              </SimpleGrid>

              <FormControl isInvalid={!!errors.assignee_id}>
                <FormLabel display="flex" alignItems="center" gap={2}>
                  <User size={14} /> Assignee
                </FormLabel>
                <Select {...register('assignee_id')} placeholder="Unassigned">
                  {members.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.name}
                    </option>
                  ))}
                </Select>
                <FormErrorMessage>
                  {errors.assignee_id?.message}
                </FormErrorMessage>
              </FormControl>
            </VStack>
          </ModalBody>
          <Box px={6} pb={6} display="flex" justifyContent="flex-end" gap={3}>
            <Button variant="ghost" onClick={handleCloseModal}>
              Cancel
            </Button>
            <Button
              type="submit"
              form="task-form"
              colorScheme="blue"
              isLoading={isSubmitting}
            >
              {editingTask ? 'Save Changes' : 'Create Task'}
            </Button>
          </Box>
        </ModalContent>
      </Modal>
    </Box>
  );
};
