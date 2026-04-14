import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import {
  Box,
  Button,
  Container,
  FormControl,
  FormLabel,
  Input,
  VStack,
  Heading,
  FormErrorMessage,
  useToast,
} from '@chakra-ui/react';
import { api } from '../../api/client';
import { CheckSquare } from 'lucide-react';

const authSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(6, 'Minimum 6 characters'),
  name: z.string().optional(),
});

type AuthValues = z.infer<typeof authSchema>;

export const AuthPage = () => {
  const [isRegister, setIsRegister] = useState(false);
  const toast = useToast();
  const navigate = useNavigate();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<AuthValues>({
    resolver: zodResolver(authSchema),
  });

  const onSubmit = async (data: AuthValues) => {
    try {
      if (isRegister && !data.name)
        return toast({ title: 'Name required', status: 'error' });
      const endpoint = isRegister ? '/auth/register' : '/auth/login';
      const res = await api.post(endpoint, data);

      if (isRegister) {
        toast({ title: 'Account created!', status: 'success' });
        setIsRegister(false);
        reset();
      } else {
        localStorage.setItem('taskflow_token', res.data.access_token);
        localStorage.setItem(
          'taskflow_user_name',
          res.data.user?.name || data.email.split('@')[0]
        );
        navigate('/projects');
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Auth failed';
      toast({
        title: 'Error',
        description: errorMessage,
        status: 'error',
      });
    }
  };

  return (
    <Container maxW="md" centerContent py={20}>
      <Box
        w="full"
        p={8}
        borderWidth={1}
        borderRadius="xl"
        boxShadow="lg"
        bg="white"
        _dark={{ bg: 'gray.800', borderColor: 'gray.700' }}
      >
        <VStack spacing={6} as="form" onSubmit={handleSubmit(onSubmit)}>
          <CheckSquare size={48} color="#3182ce" />
          <Heading size="lg">
            {isRegister ? 'Create Account' : 'Welcome Back'}
          </Heading>

          {isRegister && (
            <FormControl isInvalid={!!errors.name}>
              <FormLabel>Name</FormLabel>
              <Input {...register('name')} />
              <FormErrorMessage>{errors.name?.message}</FormErrorMessage>
            </FormControl>
          )}

          <FormControl isInvalid={!!errors.email}>
            <FormLabel>Email</FormLabel>
            <Input type="email" {...register('email')} />
            <FormErrorMessage>{errors.email?.message}</FormErrorMessage>
          </FormControl>

          <FormControl isInvalid={!!errors.password}>
            <FormLabel>Password</FormLabel>
            <Input type="password" {...register('password')} />
            <FormErrorMessage>{errors.password?.message}</FormErrorMessage>
          </FormControl>

          <Button
            type="submit"
            colorScheme="blue"
            w="full"
            isLoading={isSubmitting}
          >
            {isRegister ? 'Register' : 'Sign In'}
          </Button>

          <Button
            variant="link"
            onClick={() => {
              setIsRegister(!isRegister);
              reset();
            }}
          >
            {isRegister
              ? 'Already have an account? Login'
              : 'Need an account? Register'}
          </Button>
        </VStack>
      </Box>
    </Container>
  );
};
