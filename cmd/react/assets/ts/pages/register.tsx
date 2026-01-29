import { Form, router } from "@inertiajs/react";
import { useRef } from "react";

export default function Register(props) {
  const formRef = useRef(null);

  return (
    <Form ref={formRef} action="/users" method="post">
      {({ errors, invalid, validate, validating, setError }) => {
        return (
          <>
            <label htmlFor="name">Name:</label>
            <input
              id="name"
              name="name"
              onChange={() => {
                validate("name");
              }}
            />
            {invalid("name") && <p>{errors.name}</p>}
            <label htmlFor="email">Email:</label>
            <input id="email" name="email" onChange={() => validate("email")} />
            {invalid("email") && <p>{errors.email}</p>}
            <label htmlFor="password">Password:</label>
            <input
              id="password"
              name="password"
              onChange={() => validate("password")}
            />
            {invalid("password") && <p>{errors.password}</p>}
            <button type="submit">Create User</button>
          </>
        );
      }}
    </Form>
  );
}
