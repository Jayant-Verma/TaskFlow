import {
  Stat,
  StatLabel,
  StatNumber,
  StatGroup,
  Progress,
  Box,
  Text,
  Spinner,
  Flex,
} from '@chakra-ui/react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { ProjectStats } from '@/types';

export const ProjectStatsHeader = ({ projectId }: { projectId: string }) => {
  const { data, isLoading } = useQuery<ProjectStats>({
    queryKey: ['project-stats', projectId],
    queryFn: async () => (await api.get(`/projects/${projectId}/stats`)).data,
  });

  if (isLoading || !data) return <Spinner size="sm" mb={4} />;

  const counts = data.status_counts;
  const todo = counts.todo || 0;
  const inProgress = counts.in_progress || 0;
  const done = counts.done || 0;
  const total = todo + inProgress + done;

  const percentComplete = total > 0 ? Math.round((done / total) * 100) : 0;

  return (
    <Box
      mb={8}
      p={5}
      bg="white"
      _dark={{ bg: 'gray.800', borderColor: 'gray.700' }}
      borderRadius="xl"
      border="1px"
      borderColor="gray.200"
    >
      <StatGroup>
        <Stat>
          <StatLabel>Total Tasks</StatLabel>
          <StatNumber>{total}</StatNumber>
        </Stat>

        <Stat>
          <StatLabel>In Progress</StatLabel>
          <StatNumber color="blue.500">{inProgress}</StatNumber>
        </Stat>

        <Stat>
          <StatLabel>Completed</StatLabel>
          <StatNumber color="green.500">{done}</StatNumber>
        </Stat>
      </StatGroup>

      <Box mt={4}>
        <Flex justify="space-between" align="center" mb={1}>
          <Text fontSize="xs" fontWeight="bold" color="gray.500">
            PROJECT COMPLETION
          </Text>
          <Text fontSize="xs" fontWeight="bold" color="green.500">
            {percentComplete}%
          </Text>
        </Flex>
        <Progress
          value={percentComplete}
          size="xs"
          colorScheme="green"
          borderRadius="full"
        />
      </Box>
    </Box>
  );
};
