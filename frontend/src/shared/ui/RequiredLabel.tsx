export function RequiredLabel({ children, required = false }: { children: string; required?: boolean }) {
  return (
    <>
      {children}
      {required && <span className="text-red-600 ml-1" aria-label="obligatorio">*</span>}
    </>
  );
}
