import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';


export default defineConfig({
    integrations: [
        starlight({
            title: 'inertigo',
            social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/joetifa2003/inertigo' }],
            sidebar: [
                {
                    label: 'Introduction',
                    items: [
                        { label: 'What is Inertia?', slug: 'introduction/what-is-inertia' },
                        { label: 'Why inertigo?', slug: 'introduction/why-inertigo' },
                    ],
                },
                {
                    label: 'Getting Started',
                    items: [
                        { label: 'Installation', slug: 'getting-started/installation' },
                        { label: 'Quick Start', slug: 'getting-started/quick-start' },
                    ],
                },
                {
                    label: 'Core Concepts',
                    items: [
                        { label: 'The Inertia Instance', slug: 'core-concepts/the-inertia-instance' },
                        { label: 'Rendering Pages', slug: 'core-concepts/rendering-pages' },
                        { label: 'Middleware', slug: 'core-concepts/middleware' },
                        { label: 'Root Template', slug: 'core-concepts/root-template' },
                    ],
                },
                {
                    label: 'Props',
                    items: [
                        { label: 'Overview', slug: 'props/overview' },
                        { label: 'Value and Lazy', slug: 'props/value-and-lazy' },
                        { label: 'Deferred', slug: 'props/deferred' },
                        { label: 'Optional', slug: 'props/optional' },
                        { label: 'Always', slug: 'props/always' },
                        { label: 'Once', slug: 'props/once' },
                        { label: 'Scroll', slug: 'props/scroll' },
                    ],
                },
                {
                    label: 'Data Management',
                    items: [
                        { label: 'Shared Props', slug: 'data/shared-props' },
                        { label: 'Flash Messages', slug: 'data/flash-messages' },
                        { label: 'Validation Errors', slug: 'data/validation-errors' },
                    ],
                },
                {
                    label: 'Advanced',
                    items: [
                        { label: 'Server-Side Rendering', slug: 'advanced/server-side-rendering' },
                        { label: 'Asset Versioning', slug: 'advanced/asset-versioning' },
                        { label: 'CSRF Protection', slug: 'advanced/csrf-protection' },
                        { label: 'Precognition', slug: 'advanced/precognition' },
                        { label: 'Partial Reloads', slug: 'advanced/partial-reloads' },
                        { label: 'Render Options', slug: 'advanced/render-options' },
                    ],
                },
                {
                    label: 'Bundlers',
                    items: [
                        { label: 'Vite', slug: 'bundlers/vite' },
                        { label: 'QuickJS SSR', slug: 'bundlers/quickjs-ssr' },
                    ],
                },
            ],
        }),
    ],

    vite: {
        plugins: [],
    },
});