export interface User {
  id: string;
  name: string;
  email: string;
}
export interface Project {
  id: string;
  name: string;
  description: string;
  created_at: string;
}
export interface Task {
  id: string;
  title: string;
  description: string;
  status: 'todo' | 'in_progress' | 'done';
  priority: 'low' | 'medium' | 'high';
  project_id: string;
  assignee_id?: string;
  due_date?: string;
}
