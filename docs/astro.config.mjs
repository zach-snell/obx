// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://zach-snell.github.io',
	base: '/obsidian-go-mcp',
	integrations: [
		starlight({
			title: 'obsidian-go-mcp',
			description: 'A fast, lightweight MCP server for Obsidian vaults written in Go.',
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/zach-snell/obsidian-go-mcp' },
			],
			editLink: {
				baseUrl: 'https://github.com/zach-snell/obsidian-go-mcp/edit/main/docs/',
			},
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'getting-started/introduction' },
						{ label: 'Installation', slug: 'getting-started/installation' },
						{ label: 'Configuration', slug: 'getting-started/configuration' },
						{ label: 'Quick Start', slug: 'getting-started/quickstart' },
					],
				},
				{
					label: 'Tools Reference',
					items: [
						{ label: 'Overview', slug: 'tools/overview' },
						{ label: 'Core Operations', slug: 'tools/core' },
						{ label: 'Search', slug: 'tools/search' },
						{ label: 'Frontmatter & Fields', slug: 'tools/frontmatter' },
						{ label: 'Graph & Links', slug: 'tools/graph' },
						{ label: 'Periodic Notes', slug: 'tools/periodic' },
						{ label: 'Templates', slug: 'tools/templates' },
						{ label: 'Organization', slug: 'tools/organization' },
						{ label: 'Bulk Operations', slug: 'tools/bulk' },
						{ label: 'Canvas', slug: 'tools/canvas' },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'Usage Examples', slug: 'guides/examples' },
						{ label: 'Task Management', slug: 'guides/tasks' },
						{ label: 'Template Variables', slug: 'guides/templates' },
					],
				},
				{
					label: 'Advanced',
					items: [
						{ label: 'Security', slug: 'advanced/security' },
						{ label: 'Development', slug: 'advanced/development' },
						{ label: 'FAQ', slug: 'advanced/faq' },
					],
				},
			],
		}),
	],
});
