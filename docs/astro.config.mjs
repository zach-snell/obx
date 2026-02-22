// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://zach-snell.github.io',
	base: '/obx',
	integrations: [
		starlight({
			title: 'obx',
			description: 'A fast, lightweight MCP server for Obsidian vaults written in Go.',
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/zach-snell/obx' },
			],
			editLink: {
				baseUrl: 'https://github.com/zach-snell/obx/edit/main/docs/',
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
					label: 'CLI Commands',
					items: [
						{ label: 'Overview', slug: 'cli/overview' },
						{ label: 'obx search', slug: 'cli/search' },
						{ label: 'obx doctor', slug: 'cli/doctor' },
						{ label: 'obx daily', slug: 'cli/daily' },
						{ label: 'obx vault', slug: 'cli/vault' },
					],
				},
				{
					label: 'MCP Tool Reference',
					items: [
						{ label: 'Overview', slug: 'mcp/overview' },
						{ label: 'manage-notes', slug: 'mcp/manage-notes' },
						{ label: 'edit-note', slug: 'mcp/edit-note' },
						{ label: 'search-vault', slug: 'mcp/search-vault' },
						{ label: 'manage-periodic-notes', slug: 'mcp/manage-periodic-notes' },
						{ label: 'manage-folders', slug: 'mcp/manage-folders' },
						{ label: 'manage-frontmatter', slug: 'mcp/manage-frontmatter' },
						{ label: 'manage-tasks', slug: 'mcp/manage-tasks' },
						{ label: 'analyze-vault', slug: 'mcp/analyze-vault' },
						{ label: 'manage-canvas', slug: 'mcp/manage-canvas' },
						{ label: 'manage-mocs', slug: 'mcp/manage-mocs' },
						{ label: 'read-batch', slug: 'mcp/read-batch' },
						{ label: 'manage-links', slug: 'mcp/manage-links' },
						{ label: 'bulk-operations', slug: 'mcp/bulk-operations' },
						{ label: 'manage-templates', slug: 'mcp/manage-templates' },
						{ label: 'manage-vaults', slug: 'mcp/manage-vaults' },
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
