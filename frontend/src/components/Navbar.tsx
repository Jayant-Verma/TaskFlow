import {
  Flex,
  Button,
  Heading,
  useColorMode,
  HStack,
  IconButton,
  Text,
  Avatar,
} from '@chakra-ui/react';
import { Moon, Sun, CheckSquare, LogOut } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export const Navbar = () => {
  const { colorMode, toggleColorMode } = useColorMode();
  const navigate = useNavigate();
  const userName = localStorage.getItem('taskflow_user_name') || 'Guest';

  const handleLogout = () => {
    localStorage.removeItem('taskflow_token');
    localStorage.removeItem('taskflow_user_name');
    window.location.href = '/login';
  };

  return (
    <Flex
      px={8}
      py={4}
      align="center"
      justify="space-between"
      borderBottom="1px"
      borderColor="gray.200"
      _dark={{ borderColor: 'whiteAlpha.300', bg: 'gray.900' }}
      bg="white"
      position="sticky"
      top={0}
      zIndex={10}
    >
      <HStack
        spacing={3}
        cursor="pointer"
        onClick={() => navigate('/projects')}
      >
        <CheckSquare color="#3182ce" />
        <Heading size="md" color="blue.500">
          TaskFlow
        </Heading>
      </HStack>
      <HStack spacing={6}>
        <HStack spacing={3} display={{ base: 'none', sm: 'flex' }}>
          <Avatar size="sm" name={userName} bg="blue.500" color="white" />
          <Text fontWeight="medium" fontSize="sm">
            {userName}
          </Text>
        </HStack>
        <HStack spacing={2}>
          <IconButton
            aria-label="Toggle dark mode"
            icon={colorMode === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
            onClick={toggleColorMode}
            variant="ghost"
          />
          <Button
            size="sm"
            colorScheme="red"
            variant="ghost"
            leftIcon={<LogOut size={16} />}
            onClick={handleLogout}
          >
            Logout
          </Button>
        </HStack>
      </HStack>
    </Flex>
  );
};
