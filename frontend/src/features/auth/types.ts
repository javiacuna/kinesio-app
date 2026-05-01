export type AuthRole = "admin" | "recepcionista" | "kinesiologo" | "paciente";

export type AuthUser = {
  uid: string;
  email: string;
  role: AuthRole | "";
};

export type LoginResponse = {
  id_token: string;
  refresh_token: string;
  expires_in: string;
  uid: string;
  email: string;
};

export type LoginInput = {
  email: string;
  password: string;
};

export type PasswordResetInput = {
  email: string;
};

export type ChangePasswordInput = {
  current_password: string;
  new_password: string;
};
