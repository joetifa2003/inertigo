import { Form } from "@inertiajs/react";

export default function Register(props) {
  return (
    <Form action="/users" method="post">
      {({ errors, invalid, validate, validating }) => (
        <>
          <label htmlFor="name">Name:</label>
          <input id="name" name="name" onChange={() => validate("name")} />
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
      )}
    </Form>
  );
}
