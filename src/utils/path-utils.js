/**
 * Path utility functions for Task Master
 * Provides centralized path resolution logic for both CLI and MCP use cases
 */

import path from 'path';
import fs from 'fs';
import {
	TASKMASTER_TASKS_FILE,
	LEGACY_TASKS_FILE,
	TASKMASTER_DOCS_DIR,
	TASKMASTER_REPORTS_DIR,
	// COMPLEXITY_REPORT_FILE, // This will be constructed dynamically
	TASKMASTER_CONFIG_FILE,
	LEGACY_CONFIG_FILE
} from '../constants/paths.js';
import { getTasksPath } from '../../scripts/modules/config-manager.js';

/**
 * Find the project root directory by looking for project markers
 * @param {string} startDir - Directory to start searching from
 * @returns {string|null} - Project root path or null if not found
 */
export function findProjectRoot(startDir = process.cwd()) {
	const projectMarkers = [
		'.taskmaster',
		TASKMASTER_TASKS_FILE,
		'tasks.json',
		LEGACY_TASKS_FILE,
		'.git',
		'.svn',
		'package.json',
		'yarn.lock',
		'package-lock.json',
		'pnpm-lock.yaml'
	];

	let currentDir = path.resolve(startDir);
	const rootDir = path.parse(currentDir).root;

	while (currentDir !== rootDir) {
		// Check if current directory contains any project markers
		for (const marker of projectMarkers) {
			const markerPath = path.join(currentDir, marker);
			if (fs.existsSync(markerPath)) {
				return currentDir;
			}
		}
		currentDir = path.dirname(currentDir);
	}

	return null;
}

/**
 * Find the tasks.json file path with fallback logic
 * @param {string|null} explicitPath - Explicit path provided by user (highest priority)
 * @param {Object|null} args - Args object from MCP args (optional)
 * @param {Object|null} log - Logger object (optional)
 * @returns {string|null} - Resolved tasks.json path or null if not found
 */
export function findTasksPath(explicitPath = null, args = null, log = null) {
	const logger = log || console;

	// 1. If explicit path is provided, use it (highest priority)
	if (explicitPath) {
		const resolvedPath = path.isAbsolute(explicitPath)
			? explicitPath
			: path.resolve(process.cwd(), explicitPath);

		if (fs.existsSync(resolvedPath)) {
			logger.info?.(`Using explicit tasks path: ${resolvedPath}`);
			return resolvedPath;
		} else {
			logger.warn?.(
				`Explicit tasks path not found: ${resolvedPath}, trying fallbacks`
			);
		}
	}

	// 2. Try to get project root from args (MCP) or find it
	const projectRoot = args?.projectRoot || findProjectRoot();

	if (!projectRoot) {
		logger.warn?.('Could not determine project root directory');
		return null;
	}

	// 3. Check possible locations in order of preference
	const possiblePaths = [
		path.join(projectRoot, TASKMASTER_TASKS_FILE), // .taskmaster/tasks/tasks.json (NEW)
		path.join(projectRoot, 'tasks.json'), // tasks.json in root (LEGACY)
		path.join(projectRoot, LEGACY_TASKS_FILE) // tasks/tasks.json (LEGACY)
	];

	for (const tasksPath of possiblePaths) {
		if (fs.existsSync(tasksPath)) {
			logger.info?.(`Found tasks file at: ${tasksPath}`);

			// Issue deprecation warning for legacy paths
			if (
				tasksPath.includes('tasks/tasks.json') &&
				!tasksPath.includes('.taskmaster')
			) {
				logger.warn?.(
					`⚠️  DEPRECATION WARNING: Found tasks.json in legacy location '${tasksPath}'. Please migrate to the new .taskmaster directory structure. Run 'task-master migrate' to automatically migrate your project.`
				);
			} else if (
				tasksPath.endsWith('tasks.json') &&
				!tasksPath.includes('.taskmaster') &&
				!tasksPath.includes('tasks/')
			) {
				logger.warn?.(
					`⚠️  DEPRECATION WARNING: Found tasks.json in legacy root location '${tasksPath}'. Please migrate to the new .taskmaster directory structure. Run 'task-master migrate' to automatically migrate your project.`
				);
			}

			return tasksPath;
		}
	}

	logger.warn?.(`No tasks.json found in project: ${projectRoot}`);
	return null;
}

/**
 * Find the PRD document file path with fallback logic
 * @param {string|null} explicitPath - Explicit path provided by user (highest priority)
 * @param {Object|null} args - Args object for MCP context (optional)
 * @param {Object|null} log - Logger object (optional)
 * @returns {string|null} - Resolved PRD document path or null if not found
 */
export function findPRDPath(explicitPath = null, args = null, log = null) {
	const logger = log || console;

	// 1. If explicit path is provided, use it (highest priority)
	if (explicitPath) {
		const resolvedPath = path.isAbsolute(explicitPath)
			? explicitPath
			: path.resolve(process.cwd(), explicitPath);

		if (fs.existsSync(resolvedPath)) {
			logger.info?.(`Using explicit PRD path: ${resolvedPath}`);
			return resolvedPath;
		} else {
			logger.warn?.(
				`Explicit PRD path not found: ${resolvedPath}, trying fallbacks`
			);
		}
	}

	// 2. Try to get project root from args (MCP) or find it
	const projectRoot = args?.projectRoot || findProjectRoot();

	if (!projectRoot) {
		logger.warn?.('Could not determine project root directory');
		return null;
	}

	// 3. Check possible locations in order of preference
	const locations = [
		TASKMASTER_DOCS_DIR, // .taskmaster/docs/ (NEW)
		'scripts/', // Legacy location
		'' // Project root
	];

	const fileNames = ['PRD.md', 'prd.md', 'PRD.txt', 'prd.txt'];

	for (const location of locations) {
		for (const fileName of fileNames) {
			const prdPath = path.join(projectRoot, location, fileName);
			if (fs.existsSync(prdPath)) {
				logger.info?.(`Found PRD document at: ${prdPath}`);

				// Issue deprecation warning for legacy paths
				if (location === 'scripts/' || location === '') {
					logger.warn?.(
						`⚠️  DEPRECATION WARNING: Found PRD file in legacy location '${prdPath}'. Please migrate to .taskmaster/docs/ directory. Run 'task-master migrate' to automatically migrate your project.`
					);
				}

				return prdPath;
			}
		}
	}

	logger.warn?.(`No PRD document found in project: ${projectRoot}`);
	return null;
}

/**
 * Find the complexity report file path with fallback logic
 * @param {string|null} explicitPath - Explicit path provided by user (highest priority)
 * @param {Object|null} args - Args object for MCP context (optional)
 * @param {Object|null} log - Logger object (optional)
 * @returns {string|null} - Resolved complexity report path or null if not found
 */
export function findComplexityReportPath(
	explicitPath = null,
	args = null,
	log = null
) {
	const logger = log || console;

	// 1. If explicit path is provided, use it (highest priority)
	if (explicitPath) {
		const resolvedPath = path.isAbsolute(explicitPath)
			? explicitPath
			: path.resolve(process.cwd(), explicitPath);

		if (fs.existsSync(resolvedPath)) {
			logger.info?.(`Using explicit complexity report path: ${resolvedPath}`);
			return resolvedPath;
		} else {
			logger.warn?.(
				`Explicit complexity report path not found: ${resolvedPath}, falling back to config-based search.`
			);
			// If explicit path is given but not found, we might not want to fallback silently.
			// However, current instruction is to "try fallbacks", so we proceed.
		}
	}

	// 2. Get projectRoot
	const projectRoot = findProjectRoot(args?.projectRoot || process.cwd());

	if (!projectRoot) {
		logger.warn?.('Could not determine project root directory. Cannot find complexity report.');
		return null;
	}

	// 3. Get tasksPath from configuration
	const tasksPath = getTasksPath(projectRoot); // Assuming getTasksPath can work with just projectRoot

	if (!tasksPath) {
		logger.warn?.(`Tasks path not defined in configuration for project: ${projectRoot}. Cannot find complexity report.`);
		return null;
	}

	// 4. Resolve tasksPath relative to projectRoot to get the absolute tasksPathValue.
	// tasksPath from config might be relative to project root or absolute.
	// getTasksPath should ideally return an absolute path or a path relative to projectRoot.
	// For now, let's assume it's relative to projectRoot if not absolute.
	const tasksPathValue = path.isAbsolute(tasksPath)
		? tasksPath
		: path.resolve(projectRoot, tasksPath);

	// 5. Construct the full path to the complexity report file.
	const reportFileName = 'task-complexity-report.json'; // Standardized name
	const complexityReportPath = path.join(tasksPathValue, reportFileName);

	// 6. Check if this file exists.
	if (fs.existsSync(complexityReportPath)) {
		logger.info?.(`Found complexity report at: ${complexityReportPath}`);
		return complexityReportPath;
	} else {
		logger.warn?.(`Complexity report not found at expected path: ${complexityReportPath}`);
		// Also check for 'complexity-report.json' in the same directory as a fallback for the filename itself.
		const alternateReportFileName = 'complexity-report.json';
		const alternateComplexityReportPath = path.join(tasksPathValue, alternateReportFileName);
		if (fs.existsSync(alternateComplexityReportPath)) {
			logger.info?.(`Found complexity report with alternate name at: ${alternateComplexityReportPath}`);
			return alternateComplexityReportPath;
		}
		logger.warn?.(`Also checked for ${alternateReportFileName} in ${tasksPathValue}, not found.`);
	}

	logger.warn?.(`No complexity report found in project: ${projectRoot} based on tasksPath in configuration.`);
	return null;
}

/**
 * Resolve output path for tasks.json (create if needed)
 * @param {string|null} explicitPath - Explicit output path provided by user
 * @param {Object|null} args - Args object for MCP context (optional)
 * @param {Object|null} log - Logger object (optional)
 * @returns {string} - Resolved output path for tasks.json
 */
export function resolveTasksOutputPath(
	explicitPath = null,
	args = null,
	log = null
) {
	const logger = log || console;

	// 1. If explicit path is provided, use it
	if (explicitPath) {
		const resolvedPath = path.isAbsolute(explicitPath)
			? explicitPath
			: path.resolve(process.cwd(), explicitPath);

		logger.info?.(`Using explicit output path: ${resolvedPath}`);
		return resolvedPath;
	}

	// 2. Try to get project root from args (MCP) or find it
	const projectRoot = args?.projectRoot || findProjectRoot() || process.cwd();

	// 3. Use new .taskmaster structure by default
	const defaultPath = path.join(projectRoot, TASKMASTER_TASKS_FILE);
	logger.info?.(`Using default output path: ${defaultPath}`);

	// Ensure the directory exists
	const outputDir = path.dirname(defaultPath);
	if (!fs.existsSync(outputDir)) {
		logger.info?.(`Creating tasks directory: ${outputDir}`);
		fs.mkdirSync(outputDir, { recursive: true });
	}

	return defaultPath;
}

/**
 * Resolve output path for complexity report (create if needed)
 * @param {string|null} explicitPath - Explicit output path provided by user
 * @param {Object|null} args - Args object for MCP context (optional)
 * @param {Object|null} log - Logger object (optional)
 * @returns {string} - Resolved output path for complexity report
 */
export function resolveComplexityReportOutputPath(
	explicitPath = null,
	args = null,
	log = null
) {
	const logger = log || console;

	// 1. If explicit path is provided, use it
	if (explicitPath) {
		const resolvedPath = path.isAbsolute(explicitPath)
			? explicitPath
			: path.resolve(process.cwd(), explicitPath);

		logger.info?.(
			`Using explicit complexity report output path: ${resolvedPath}`
		);
		return resolvedPath;
	}

	// 2. Try to get project root from args (MCP) or find it
	const projectRoot = args?.projectRoot || findProjectRoot() || process.cwd();

	// 3. Use new .taskmaster structure by default
	const defaultPath = path.join(projectRoot, COMPLEXITY_REPORT_FILE);
	logger.info?.(`Using default complexity report output path: ${defaultPath}`);

	// Ensure the directory exists
	const outputDir = path.dirname(defaultPath);
	if (!fs.existsSync(outputDir)) {
		logger.info?.(`Creating reports directory: ${outputDir}`);
		fs.mkdirSync(outputDir, { recursive: true });
	}

	return defaultPath;
}

/**
 * Find the configuration file path with fallback logic
 * @param {string|null} explicitPath - Explicit path provided by user (highest priority)
 * @param {Object|null} args - Args object for MCP context (optional)
 * @param {Object|null} log - Logger object (optional)
 * @returns {string|null} - Resolved config file path or null if not found
 */
export function findConfigPath(explicitPath = null, args = null, log = null) {
	const logger = log || console;

	// 1. If explicit path is provided, use it (highest priority)
	if (explicitPath) {
		const resolvedPath = path.isAbsolute(explicitPath)
			? explicitPath
			: path.resolve(process.cwd(), explicitPath);

		if (fs.existsSync(resolvedPath)) {
			logger.info?.(`Using explicit config path: ${resolvedPath}`);
			return resolvedPath;
		} else {
			logger.warn?.(
				`Explicit config path not found: ${resolvedPath}, trying fallbacks`
			);
		}
	}

	// 2. Try to get project root from args (MCP) or find it
	const projectRoot = args?.projectRoot || findProjectRoot();

	if (!projectRoot) {
		logger.warn?.('Could not determine project root directory');
		return null;
	}

	// 3. Check possible locations in order of preference
	const possiblePaths = [
		path.join(projectRoot, TASKMASTER_CONFIG_FILE), // NEW location
		path.join(projectRoot, LEGACY_CONFIG_FILE) // LEGACY location
	];

	for (const configPath of possiblePaths) {
		if (fs.existsSync(configPath)) {
			try {
				logger.info?.(`Found config file at: ${configPath}`);
			} catch (error) {
				// Silently handle logging errors during testing
			}

			// Issue deprecation warning for legacy paths
			if (configPath?.endsWith(LEGACY_CONFIG_FILE)) {
				logger.warn?.(
					`⚠️  DEPRECATION WARNING: Found configuration in legacy location '${configPath}'. Please migrate to .taskmaster/config.json. Run 'task-master migrate' to automatically migrate your project.`
				);
			}

			return configPath;
		}
	}

	logger.warn?.(`No configuration file found in project: ${projectRoot}`);
	return null;
}
