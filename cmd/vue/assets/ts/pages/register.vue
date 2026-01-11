<template>
  <Form action="/users" method="post">
    <template #default="{ errors, invalid, validate, validating, setError }">
      <label for="name">Name:</label>
      <input id="name" name="name" @input="validate('name')" />
      <p v-if="invalid('name')">{{ errors.name }}</p>

      <label for="email">Email:</label>
      <input id="email" name="email" @input="validate('email')" />
      <p v-if="invalid('email')">{{ errors.email }}</p>

      <label for="password">Password:</label>
      <input id="password" name="password" @input="validate('password')" />
      <p v-if="invalid('password')">{{ errors.password }}</p>

      <button type="submit">Create User</button>
    </template>
  </Form>
</template>

<script setup>
import { Form } from "@inertiajs/vue3";
</script>
