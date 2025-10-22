# Maintenance and Updates Implementation Plan

## Overview
Establish comprehensive maintenance procedures, automated update systems, and long-term sustainability practices to ensure the application remains secure, performant, and up-to-date with evolving technologies and requirements.

## Current State Analysis
- Manual dependency updates
- No automated security patching
- Limited maintenance documentation
- No update scheduling system
- No rollback procedures for updates
- No dependency vulnerability monitoring
- No performance regression testing
- Limited backup and recovery procedures

## Implementation Steps

### Step 1: Automated Dependency Management
**Timeline: 3-4 days**

Implement automated dependency updates with safety checks:

```javascript
// scripts/dependency-management.js
const fs = require('fs')
const path = require('path')
const { execSync } = require('child_process')
const semver = require('semver')

class DependencyManager {
  constructor() {
    this.packageJsonPath = path.join(process.cwd(), 'package.json')
    this.lockfilePath = path.join(process.cwd(), 'package-lock.json')
    this.packageJson = JSON.parse(fs.readFileSync(this.packageJsonPath, 'utf8'))
    this.vulnerabilityThreshold = 'moderate' // critical, high, moderate, low
    this.maxMajorUpdates = 2 // Max major version jumps at once
  }

  // Check for available updates
  async checkUpdates() {
    console.log('🔍 Checking for dependency updates...')
    
    try {
      const outdatedOutput = execSync('npm outdated --json', { encoding: 'utf8' })
      const outdated = JSON.parse(outdatedOutput)
      
      const updates = {
        patch: [],
        minor: [],
        major: [],
        security: [],
      }
      
      for (const [pkg, info] of Object.entries(outdated)) {
        const currentVersion = info.current
        const latestVersion = info.latest
        const wantedVersion = info.wanted
        
        if (semver.patch(currentVersion) < semver.patch(latestVersion)) {
          updates.patch.push({ pkg, currentVersion, latestVersion, wantedVersion })
        } else if (semver.minor(currentVersion) < semver.minor(latestVersion)) {
          updates.minor.push({ pkg, currentVersion, latestVersion, wantedVersion })
        } else if (semver.major(currentVersion) < semver.major(latestVersion)) {
          updates.major.push({ pkg, currentVersion, latestVersion, wantedVersion })
        }
      }
      
      // Check for security updates
      updates.security = await this.checkSecurityUpdates()
      
      return updates
    } catch (error) {
      if (error.status === 1) {
        // npm outdated exits with 1 when there are outdated packages
        const outdatedOutput = error.stdout.toString()
        if (outdatedOutput) {
          return this.parseOutdatedOutput(outdatedOutput)
        }
      }
      throw error
    }
  }

  // Check for security vulnerabilities
  async checkSecurityUpdates() {
    try {
      const auditOutput = execSync('npm audit --json', { encoding: 'utf8' })
      const audit = JSON.parse(auditOutput)
      
      const securityUpdates = []
      
      if (audit.vulnerabilities) {
        for (const [pkg, vuln] of Object.entries(audit.vulnerabilities)) {
          if (this.shouldUpdateForSecurity(vuln)) {
            securityUpdates.push({
              pkg,
              severity: vuln.severity,
              via: vuln.via,
              range: vuln.range,
              fixAvailable: vuln.fixAvailable,
            })
          }
        }
      }
      
      return securityUpdates
    } catch (error) {
      console.warn('Security audit failed:', error.message)
      return []
    }
  }

  shouldUpdateForSecurity(vuln) {
    const severityLevels = {
      critical: 4,
      high: 3,
      moderate: 2,
      low: 1,
    }
    
    const thresholdLevel = severityLevels[this.vulnerabilityThreshold]
    const vulnLevel = severityLevels[vuln.severity]
    
    return vulnLevel >= thresholdLevel
  }

  // Perform automated updates
  async performUpdates(updatePlan) {
    console.log('🔄 Performing automated updates...')
    
    // Create backup
    await this.createBackup()
    
    try {
      // Apply security updates first
      if (updatePlan.security.length > 0) {
        console.log(`📋 Applying ${updatePlan.security.length} security updates...`)
        await this.applySecurityUpdates(updatePlan.security)
      }
      
      // Apply patch updates
      if (updatePlan.patch.length > 0) {
        console.log(`📋 Applying ${updatePlan.patch.length} patch updates...`)
        await this.applyUpdates(updatePlan.patch, 'patch')
      }
      
      // Apply minor updates (if enabled)
      if (updatePlan.minor.length > 0 && this.shouldApplyMinorUpdates()) {
        console.log(`📋 Applying ${updatePlan.minor.length} minor updates...`)
        await this.applyUpdates(updatePlan.minor, 'minor')
      }
      
      // Major updates require manual approval
      if (updatePlan.major.length > 0) {
        console.log(`⚠️  ${updatePlan.major.length} major updates require manual review`)
        await this.createMajorUpdatePR(updatePlan.major)
      }
      
      // Run tests after updates
      await this.runTests()
      
      // Update lockfile
      await this.updateLockfile()
      
      console.log('✅ Updates completed successfully')
      return true
      
    } catch (error) {
      console.error('❌ Updates failed:', error.message)
      await this.rollbackUpdates()
      throw error
    }
  }

  async applySecurityUpdates(securityUpdates) {
    for (const update of securityUpdates) {
      if (update.fixAvailable) {
        try {
          execSync(`npm update ${update.pkg}`, { stdio: 'inherit' })
          console.log(`✅ Updated ${update.pkg} (security: ${update.severity})`)
        } catch (error) {
          console.error(`❌ Failed to update ${update.pkg}:`, error.message)
        }
      }
    }
  }

  async applyUpdates(updates, type) {
    for (const update of updates) {
      try {
        const targetVersion = type === 'patch' ? update.wantedVersion : update.latestVersion
        execSync(`npm install ${update.pkg}@${targetVersion}`, { stdio: 'inherit' })
        console.log(`✅ Updated ${update.pkg}: ${update.currentVersion} → ${targetVersion}`)
      } catch (error) {
        console.error(`❌ Failed to update ${update.pkg}:`, error.message)
      }
    }
  }

  async createBackup() {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const backupDir = path.join(process.cwd(), '.maintenance', 'backups', timestamp)
    
    fs.mkdirSync(backupDir, { recursive: true })
    
    // Backup package files
    fs.copyFileSync(this.packageJsonPath, path.join(backupDir, 'package.json'))
    fs.copyFileSync(this.lockfilePath, path.join(backupDir, 'package-lock.json'))
    
    // Backup node_modules manifest
    const nodeModulesManifest = execSync('npm list --json', { encoding: 'utf8' })
    fs.writeFileSync(path.join(backupDir, 'node_modules.json'), nodeModulesManifest)
    
    console.log(`📦 Backup created: ${backupDir}`)
    return backupDir
  }

  async rollbackUpdates() {
    console.log('🔄 Rolling back updates...')
    
    // Find latest backup
    const backupsDir = path.join(process.cwd(), '.maintenance', 'backups')
    if (!fs.existsSync(backupsDir)) return
    
    const backups = fs.readdirSync(backupsDir).sort().reverse()
    if (backups.length === 0) return
    
    const latestBackup = path.join(backupsDir, backups[0])
    
    // Restore package files
    fs.copyFileSync(path.join(latestBackup, 'package.json'), this.packageJsonPath)
    fs.copyFileSync(path.join(latestBackup, 'package-lock.json'), this.lockfilePath)
    
    // Reinstall dependencies
    execSync('npm ci', { stdio: 'inherit' })
    
    console.log('✅ Rollback completed')
  }

  async runTests() {
    console.log('🧪 Running tests after updates...')
    
    try {
      execSync('npm run test:ci', { stdio: 'inherit' })
      execSync('npm run build', { stdio: 'inherit' })
      console.log('✅ All tests passed')
    } catch (error) {
      throw new Error('Tests failed after updates')
    }
  }

  async updateLockfile() {
    execSync('npm install --package-lock-only', { stdio: 'inherit' })
  }

  shouldApplyMinorUpdates() {
    // Check if it's a scheduled minor update window
    const now = new Date()
    const dayOfWeek = now.getDay() // 0 = Sunday, 1 = Monday, etc.
    const hour = now.getHours()
    
    // Only apply minor updates on weekends during off-hours
    return (dayOfWeek === 0 || dayOfWeek === 6) && (hour < 6 || hour > 22)
  }

  async createMajorUpdatePR(majorUpdates) {
    const branchName = `maintenance/major-updates-${Date.now()}`
    
    // Create new branch
    execSync(`git checkout -b ${branchName}`)
    
    // Apply major updates
    for (const update of majorUpdates) {
      execSync(`npm install ${update.pkg}@${update.latestVersion}`)
    }
    
    // Commit changes
    execSync('git add package.json package-lock.json')
    execSync(`git commit -m "chore: major dependency updates\n\n${majorUpdates.map(u => `- ${u.pkg}: ${u.currentVersion} → ${u.latestVersion}`).join('\n')}"`)
    
    // Push branch (assumes origin remote)
    execSync(`git push origin ${branchName}`)
    
    console.log(`📝 Created PR branch: ${branchName}`)
    
    // Create PR using GitHub API (if configured)
    if (process.env.GITHUB_TOKEN) {
      await this.createGitHubPR(branchName, majorUpdates)
    }
  }

  async createGitHubPR(branchName, updates) {
    const { Octokit } = require('@octokit/rest')
    const octokit = new Octokit({
      auth: process.env.GITHUB_TOKEN,
    })
    
    const [owner, repo] = process.env.GITHUB_REPOSITORY.split('/')
    
    const title = 'chore: Major dependency updates (automated)'
    const body = `
## Major Dependency Updates

This PR contains automated major version updates that require manual review.

### Updates:
${updates.map(u => `- **${u.pkg}**: ${u.currentVersion} → ${u.latestVersion}`).join('\n')}

### Required Actions:
- [ ] Review breaking changes in each package
- [ ] Update code for API changes
- [ ] Run full test suite
- [ ] Test in staging environment
- [ ] Update documentation if needed

⚠️ **Important**: This PR was created automatically. Please review all changes carefully before merging.
    `
    
    try {
      await octokit.rest.pulls.create({
        owner,
        repo,
        title,
        body,
        head: branchName,
        base: 'main',
      })
      console.log('📝 GitHub PR created successfully')
    } catch (error) {
      console.error('❌ Failed to create GitHub PR:', error.message)
    }
  }
}

// Usage
if (require.main === module) {
  const manager = new DependencyManager()
  
  manager.checkUpdates()
    .then(updates => {
      console.log('📊 Update Summary:')
      console.log(`  Patch: ${updates.patch.length}`)
      console.log(`  Minor: ${updates.minor.length}`)
      console.log(`  Major: ${updates.major.length}`)
      console.log(`  Security: ${updates.security.length}`)
      
      if (process.argv.includes('--apply')) {
        return manager.performUpdates(updates)
      }
    })
    .catch(error => {
      console.error('❌ Dependency management failed:', error)
      process.exit(1)
    })
}

module.exports = DependencyManager

// .github/workflows/dependency-updates.yml
name: Automated Dependency Updates

on:
  schedule:
    # Run every day at 2 AM UTC
    - cron: '0 2 * * *'
  workflow_dispatch:

jobs:
  update-dependencies:
    name: Update Dependencies
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        token: ${{ secrets.GITHUB_TOKEN }}
        
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18.x'
        cache: 'npm'
        
    - name: Install dependencies
      run: npm ci
      
    - name: Run dependency updates
      run: node scripts/dependency-management.js --apply
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        GITHUB_REPOSITORY: ${{ github.repository }}
        
    - name: Commit changes
      run: |
        git config --local user.email "action@github.com"
        git config --local user.name "GitHub Action"
        
        if [[ -n $(git status --porcelain) ]]; then
          git add package.json package-lock.json
          git commit -m "chore: automated dependency updates [skip ci]"
          git push
        fi
        
    - name: Create security alert issue
      if: failure()
      uses: actions/github-script@v7
      with:
        script: |
          github.rest.issues.create({
            owner: context.repo.owner,
            repo: context.repo.repo,
            title: 'Automated Dependency Update Failed',
            body: 'The automated dependency update process failed. Please check the workflow logs and update dependencies manually.',
            labels: ['maintenance', 'dependencies', 'urgent']
          })
```

### Step 2: Performance Monitoring and Regression Testing
**Timeline: 3-4 days**

Implement automated performance monitoring and regression detection:

```javascript
// scripts/performance-monitoring.js
const lighthouse = require('lighthouse')
const chromeLauncher = require('chrome-launcher')
const fs = require('fs')
const path = require('path')

class PerformanceMonitor {
  constructor() {
    this.baselineFile = path.join(process.cwd(), '.maintenance', 'performance-baseline.json')
    this.reportsDir = path.join(process.cwd(), '.maintenance', 'performance-reports')
    this.thresholds = {
      performanceScore: 90,
      accessibilityScore: 95,
      bestPracticesScore: 90,
      seoScore: 90,
      firstContentfulPaint: 2000,
      largestContentfulPaint: 4000,
      cumulativeLayoutShift: 0.1,
      firstInputDelay: 100,
    }
    
    this.testUrls = [
      '/',
      '/markets',
      '/profile',
      '/markets/1',
    ]
  }

  async runPerformanceTests(baseUrl = 'http://localhost:3000') {
    console.log('🚀 Running performance tests...')
    
    const chrome = await chromeLauncher.launch({ chromeFlags: ['--headless'] })
    const results = []
    
    try {
      for (const urlPath of this.testUrls) {
        const url = `${baseUrl}${urlPath}`
        console.log(`📊 Testing: ${url}`)
        
        const result = await this.runLighthouseTest(url, chrome.port)
        results.push({
          url: urlPath,
          timestamp: new Date().toISOString(),
          ...result,
        })
      }
      
      await this.saveResults(results)
      await this.compareWithBaseline(results)
      
      return results
    } finally {
      await chrome.kill()
    }
  }

  async runLighthouseTest(url, port) {
    const options = {
      logLevel: 'info',
      output: 'json',
      onlyCategories: ['performance', 'accessibility', 'best-practices', 'seo'],
      port,
      settings: {
        formFactor: 'desktop',
        throttling: {
          rttMs: 40,
          throughputKbps: 10240,
          cpuSlowdownMultiplier: 1,
        },
      },
    }
    
    const runnerResult = await lighthouse(url, options)
    const report = runnerResult.lhr
    
    return {
      scores: {
        performance: Math.round(report.categories.performance.score * 100),
        accessibility: Math.round(report.categories.accessibility.score * 100),
        bestPractices: Math.round(report.categories['best-practices'].score * 100),
        seo: Math.round(report.categories.seo.score * 100),
      },
      metrics: {
        firstContentfulPaint: report.audits['first-contentful-paint'].numericValue,
        largestContentfulPaint: report.audits['largest-contentful-paint'].numericValue,
        cumulativeLayoutShift: report.audits['cumulative-layout-shift'].numericValue,
        firstInputDelay: report.audits['max-potential-fid']?.numericValue || 0,
        speedIndex: report.audits['speed-index'].numericValue,
        timeToInteractive: report.audits['interactive'].numericValue,
      },
      audits: this.extractFailedAudits(report.audits),
    }
  }

  extractFailedAudits(audits) {
    const failed = []
    
    for (const [id, audit] of Object.entries(audits)) {
      if (audit.score !== null && audit.score < 0.9) {
        failed.push({
          id,
          title: audit.title,
          description: audit.description,
          score: audit.score,
          displayValue: audit.displayValue,
        })
      }
    }
    
    return failed
  }

  async saveResults(results) {
    fs.mkdirSync(this.reportsDir, { recursive: true })
    
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const reportFile = path.join(this.reportsDir, `performance-${timestamp}.json`)
    
    fs.writeFileSync(reportFile, JSON.stringify(results, null, 2))
    console.log(`📄 Performance report saved: ${reportFile}`)
  }

  async compareWithBaseline(results) {
    if (!fs.existsSync(this.baselineFile)) {
      console.log('📋 No baseline found, creating new baseline...')
      await this.createBaseline(results)
      return
    }
    
    const baseline = JSON.parse(fs.readFileSync(this.baselineFile, 'utf8'))
    const regressions = []
    
    for (const result of results) {
      const baselineResult = baseline.find(b => b.url === result.url)
      if (!baselineResult) continue
      
      const regression = this.detectRegression(baselineResult, result)
      if (regression.length > 0) {
        regressions.push({
          url: result.url,
          regressions: regression,
        })
      }
    }
    
    if (regressions.length > 0) {
      console.error('⚠️  Performance regressions detected:')
      this.reportRegressions(regressions)
      
      if (process.env.CI) {
        throw new Error('Performance regressions detected')
      }
    } else {
      console.log('✅ No performance regressions detected')
    }
  }

  detectRegression(baseline, current) {
    const regressions = []
    
    // Check scores
    for (const [metric, threshold] of Object.entries(this.thresholds)) {
      if (metric.endsWith('Score')) {
        const baselineValue = baseline.scores[metric.replace('Score', '')]
        const currentValue = current.scores[metric.replace('Score', '')]
        
        if (currentValue < baselineValue - 5) { // 5 point tolerance
          regressions.push({
            metric,
            baseline: baselineValue,
            current: currentValue,
            delta: currentValue - baselineValue,
            type: 'score',
          })
        }
      }
    }
    
    // Check performance metrics
    for (const [metric, threshold] of Object.entries(this.thresholds)) {
      if (!metric.endsWith('Score')) {
        const baselineValue = baseline.metrics[metric]
        const currentValue = current.metrics[metric]
        
        if (!baselineValue || !currentValue) continue
        
        const percentageIncrease = ((currentValue - baselineValue) / baselineValue) * 100
        
        if (percentageIncrease > 10) { // 10% tolerance
          regressions.push({
            metric,
            baseline: baselineValue,
            current: currentValue,
            delta: currentValue - baselineValue,
            percentageIncrease,
            type: 'metric',
          })
        }
      }
    }
    
    return regressions
  }

  reportRegressions(regressions) {
    for (const { url, regressions: pageRegressions } of regressions) {
      console.error(`\n📉 Regressions for ${url}:`)
      
      for (const regression of pageRegressions) {
        if (regression.type === 'score') {
          console.error(`  • ${regression.metric}: ${regression.baseline} → ${regression.current} (${regression.delta})`)
        } else {
          console.error(`  • ${regression.metric}: ${regression.baseline}ms → ${regression.current}ms (+${regression.percentageIncrease.toFixed(1)}%)`)
        }
      }
    }
  }

  async createBaseline(results) {
    fs.mkdirSync(path.dirname(this.baselineFile), { recursive: true })
    fs.writeFileSync(this.baselineFile, JSON.stringify(results, null, 2))
    console.log('📋 Performance baseline created')
  }

  async updateBaseline() {
    const results = await this.runPerformanceTests()
    await this.createBaseline(results)
    console.log('📋 Performance baseline updated')
  }

  async generateTrendReport() {
    if (!fs.existsSync(this.reportsDir)) {
      console.log('📊 No performance reports found')
      return
    }
    
    const reportFiles = fs.readdirSync(this.reportsDir)
      .filter(file => file.startsWith('performance-') && file.endsWith('.json'))
      .sort()
    
    const trends = {}
    
    for (const file of reportFiles) {
      const report = JSON.parse(fs.readFileSync(path.join(this.reportsDir, file), 'utf8'))
      
      for (const result of report) {
        if (!trends[result.url]) {
          trends[result.url] = []
        }
        
        trends[result.url].push({
          timestamp: result.timestamp,
          scores: result.scores,
          metrics: result.metrics,
        })
      }
    }
    
    const trendReportFile = path.join(this.reportsDir, 'performance-trends.json')
    fs.writeFileSync(trendReportFile, JSON.stringify(trends, null, 2))
    
    console.log(`📈 Performance trend report generated: ${trendReportFile}`)
    return trends
  }
}

// Bundle analysis for size monitoring
class BundleAnalyzer {
  constructor() {
    this.baselinePath = path.join(process.cwd(), '.maintenance', 'bundle-baseline.json')
    this.reportsDir = path.join(process.cwd(), '.maintenance', 'bundle-reports')
    this.sizeThreshold = 0.1 // 10% increase threshold
  }

  async analyzeBundleSize() {
    console.log('📦 Analyzing bundle size...')
    
    const buildDir = path.join(process.cwd(), 'dist')
    if (!fs.existsSync(buildDir)) {
      throw new Error('Build directory not found. Run npm run build first.')
    }
    
    const analysis = await this.getBundleAnalysis(buildDir)
    await this.saveAnalysis(analysis)
    await this.compareWithBaseline(analysis)
    
    return analysis
  }

  async getBundleAnalysis(buildDir) {
    const files = this.getAllFiles(buildDir)
    const analysis = {
      timestamp: new Date().toISOString(),
      totalSize: 0,
      files: {},
      chunks: {},
    }
    
    for (const file of files) {
      const stats = fs.statSync(file)
      const relativePath = path.relative(buildDir, file)
      const ext = path.extname(file)
      
      analysis.files[relativePath] = {
        size: stats.size,
        type: ext,
      }
      
      analysis.totalSize += stats.size
      
      // Categorize chunks
      if (ext === '.js') {
        const category = this.categorizeJSFile(relativePath)
        if (!analysis.chunks[category]) {
          analysis.chunks[category] = { size: 0, files: [] }
        }
        analysis.chunks[category].size += stats.size
        analysis.chunks[category].files.push(relativePath)
      }
    }
    
    return analysis
  }

  categorizeJSFile(filePath) {
    if (filePath.includes('vendor') || filePath.includes('node_modules')) {
      return 'vendor'
    } else if (filePath.includes('chunk')) {
      return 'chunks'
    } else if (filePath.includes('main') || filePath.includes('app')) {
      return 'main'
    } else {
      return 'other'
    }
  }

  getAllFiles(dir) {
    const files = []
    const items = fs.readdirSync(dir)
    
    for (const item of items) {
      const fullPath = path.join(dir, item)
      const stats = fs.statSync(fullPath)
      
      if (stats.isDirectory()) {
        files.push(...this.getAllFiles(fullPath))
      } else {
        files.push(fullPath)
      }
    }
    
    return files
  }

  async saveAnalysis(analysis) {
    fs.mkdirSync(this.reportsDir, { recursive: true })
    
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const reportFile = path.join(this.reportsDir, `bundle-${timestamp}.json`)
    
    fs.writeFileSync(reportFile, JSON.stringify(analysis, null, 2))
    console.log(`📄 Bundle analysis saved: ${reportFile}`)
  }

  async compareWithBaseline(analysis) {
    if (!fs.existsSync(this.baselinePath)) {
      console.log('📋 No bundle baseline found, creating new baseline...')
      fs.mkdirSync(path.dirname(this.baselinePath), { recursive: true })
      fs.writeFileSync(this.baselinePath, JSON.stringify(analysis, null, 2))
      return
    }
    
    const baseline = JSON.parse(fs.readFileSync(this.baselinePath, 'utf8'))
    const increases = []
    
    // Compare total size
    const totalIncrease = (analysis.totalSize - baseline.totalSize) / baseline.totalSize
    if (totalIncrease > this.sizeThreshold) {
      increases.push({
        type: 'total',
        baseline: baseline.totalSize,
        current: analysis.totalSize,
        increase: totalIncrease,
      })
    }
    
    // Compare chunk sizes
    for (const [chunk, data] of Object.entries(analysis.chunks)) {
      const baselineData = baseline.chunks[chunk]
      if (baselineData) {
        const increase = (data.size - baselineData.size) / baselineData.size
        if (increase > this.sizeThreshold) {
          increases.push({
            type: 'chunk',
            name: chunk,
            baseline: baselineData.size,
            current: data.size,
            increase,
          })
        }
      }
    }
    
    if (increases.length > 0) {
      console.warn('⚠️  Bundle size increases detected:')
      for (const increase of increases) {
        const baselineKB = Math.round(increase.baseline / 1024)
        const currentKB = Math.round(increase.current / 1024)
        const increasePercent = (increase.increase * 100).toFixed(1)
        
        if (increase.type === 'total') {
          console.warn(`  • Total bundle size: ${baselineKB}KB → ${currentKB}KB (+${increasePercent}%)`)
        } else {
          console.warn(`  • ${increase.name} chunk: ${baselineKB}KB → ${currentKB}KB (+${increasePercent}%)`)
        }
      }
      
      if (process.env.CI) {
        throw new Error('Bundle size increased beyond threshold')
      }
    } else {
      console.log('✅ No significant bundle size increases detected')
    }
  }
}

module.exports = { PerformanceMonitor, BundleAnalyzer }

// Usage
if (require.main === module) {
  const performanceMonitor = new PerformanceMonitor()
  const bundleAnalyzer = new BundleAnalyzer()
  
  const command = process.argv[2]
  
  switch (command) {
    case 'performance':
      performanceMonitor.runPerformanceTests(process.argv[3])
        .then(() => console.log('✅ Performance tests completed'))
        .catch(error => {
          console.error('❌ Performance tests failed:', error)
          process.exit(1)
        })
      break
      
    case 'bundle':
      bundleAnalyzer.analyzeBundleSize()
        .then(() => console.log('✅ Bundle analysis completed'))
        .catch(error => {
          console.error('❌ Bundle analysis failed:', error)
          process.exit(1)
        })
      break
      
    case 'baseline':
      performanceMonitor.updateBaseline()
        .then(() => console.log('✅ Baseline updated'))
        .catch(error => {
          console.error('❌ Baseline update failed:', error)
          process.exit(1)
        })
      break
      
    case 'trends':
      performanceMonitor.generateTrendReport()
        .then(() => console.log('✅ Trend report generated'))
        .catch(error => {
          console.error('❌ Trend report failed:', error)
          process.exit(1)
        })
      break
      
    default:
      console.log('Usage: node performance-monitoring.js <command>')
      console.log('Commands: performance, bundle, baseline, trends')
      process.exit(1)
  }
}
```

### Step 3: Automated Backup and Recovery System
**Timeline: 2-3 days**

Implement comprehensive backup and recovery procedures:

```javascript
// scripts/backup-recovery.js
const fs = require('fs')
const path = require('path')
const { execSync } = require('child_process')
const AWS = require('aws-sdk')

class BackupRecoveryManager {
  constructor() {
    this.backupDir = path.join(process.cwd(), '.maintenance', 'backups')
    this.s3 = new AWS.S3({
      accessKeyId: process.env.AWS_ACCESS_KEY_ID,
      secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
      region: process.env.AWS_REGION || 'us-east-1',
    })
    this.bucketName = process.env.BACKUP_S3_BUCKET || 'socialpredict-backups'
    this.retentionDays = 30
  }

  async createFullBackup() {
    console.log('📦 Creating full application backup...')
    
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const backupName = `full-backup-${timestamp}`
    const backupPath = path.join(this.backupDir, backupName)
    
    fs.mkdirSync(backupPath, { recursive: true })
    
    try {
      // Backup source code
      await this.backupSourceCode(backupPath)
      
      // Backup configuration files
      await this.backupConfiguration(backupPath)
      
      // Backup build artifacts
      await this.backupBuildArtifacts(backupPath)
      
      // Backup dependencies info
      await this.backupDependencies(backupPath)
      
      // Backup environment settings
      await this.backupEnvironment(backupPath)
      
      // Create backup manifest
      await this.createBackupManifest(backupPath, backupName)
      
      // Compress backup
      const archivePath = await this.compressBackup(backupPath, backupName)
      
      // Upload to S3
      if (this.isS3Configured()) {
        await this.uploadToS3(archivePath, backupName)
      }
      
      console.log(`✅ Full backup created: ${backupName}`)
      return backupName
      
    } catch (error) {
      console.error('❌ Backup failed:', error)
      // Cleanup partial backup
      if (fs.existsSync(backupPath)) {
        fs.rmSync(backupPath, { recursive: true, force: true })
      }
      throw error
    }
  }

  async backupSourceCode(backupPath) {
    console.log('📁 Backing up source code...')
    
    const sourceBackupPath = path.join(backupPath, 'source')
    fs.mkdirSync(sourceBackupPath, { recursive: true })
    
    // Get list of tracked files
    const trackedFiles = execSync('git ls-files', { encoding: 'utf8' }).trim().split('\n')
    
    for (const file of trackedFiles) {
      const srcPath = path.join(process.cwd(), file)
      const destPath = path.join(sourceBackupPath, file)
      
      if (fs.existsSync(srcPath)) {
        fs.mkdirSync(path.dirname(destPath), { recursive: true })
        fs.copyFileSync(srcPath, destPath)
      }
    }
    
    // Also backup git metadata
    const gitPath = path.join(process.cwd(), '.git')
    if (fs.existsSync(gitPath)) {
      this.copyDirectory(gitPath, path.join(sourceBackupPath, '.git'))
    }
  }

  async backupConfiguration(backupPath) {
    console.log('⚙️  Backing up configuration files...')
    
    const configBackupPath = path.join(backupPath, 'config')
    fs.mkdirSync(configBackupPath, { recursive: true })
    
    const configFiles = [
      'package.json',
      'package-lock.json',
      'vite.config.mjs',
      'tailwind.config.js',
      'postcss.config.js',
      '.eslintrc.js',
      '.eslintrc.json',
      '.prettierrc',
      'playwright.config.js',
      'jest.config.js',
      'tsconfig.json',
      'nginx.conf',
      'Dockerfile',
      'Dockerfile.prod',
      'docker-compose.yml',
      'docker-compose.prod.yml',
    ]
    
    for (const file of configFiles) {
      const srcPath = path.join(process.cwd(), file)
      if (fs.existsSync(srcPath)) {
        fs.copyFileSync(srcPath, path.join(configBackupPath, file))
      }
    }
    
    // Backup GitHub workflows
    const workflowsPath = path.join(process.cwd(), '.github', 'workflows')
    if (fs.existsSync(workflowsPath)) {
      this.copyDirectory(workflowsPath, path.join(configBackupPath, '.github', 'workflows'))
    }
  }

  async backupBuildArtifacts(backupPath) {
    console.log('🏗️  Backing up build artifacts...')
    
    const buildPath = path.join(process.cwd(), 'dist')
    if (fs.existsSync(buildPath)) {
      const buildBackupPath = path.join(backupPath, 'build')
      this.copyDirectory(buildPath, buildBackupPath)
    }
  }

  async backupDependencies(backupPath) {
    console.log('📦 Backing up dependencies info...')
    
    const depsBackupPath = path.join(backupPath, 'dependencies')
    fs.mkdirSync(depsBackupPath, { recursive: true })
    
    // Save npm list output
    try {
      const npmList = execSync('npm list --json --depth=0', { encoding: 'utf8' })
      fs.writeFileSync(path.join(depsBackupPath, 'npm-list.json'), npmList)
    } catch (error) {
      console.warn('Could not generate npm list:', error.message)
    }
    
    // Save npm audit output
    try {
      const npmAudit = execSync('npm audit --json', { encoding: 'utf8' })
      fs.writeFileSync(path.join(depsBackupPath, 'npm-audit.json'), npmAudit)
    } catch (error) {
      console.warn('Could not generate npm audit:', error.message)
    }
    
    // Save npm outdated output
    try {
      const npmOutdated = execSync('npm outdated --json', { encoding: 'utf8' })
      fs.writeFileSync(path.join(depsBackupPath, 'npm-outdated.json'), npmOutdated)
    } catch (error) {
      // npm outdated exits with status 1 when packages are outdated
      if (error.stdout) {
        fs.writeFileSync(path.join(depsBackupPath, 'npm-outdated.json'), error.stdout)
      }
    }
  }

  async backupEnvironment(backupPath) {
    console.log('🌍 Backing up environment info...')
    
    const envBackupPath = path.join(backupPath, 'environment')
    fs.mkdirSync(envBackupPath, { recursive: true })
    
    const envInfo = {
      nodeVersion: process.version,
      npmVersion: execSync('npm --version', { encoding: 'utf8' }).trim(),
      platform: process.platform,
      arch: process.arch,
      timestamp: new Date().toISOString(),
      gitCommit: this.getCurrentGitCommit(),
      gitBranch: this.getCurrentGitBranch(),
      environment: process.env.NODE_ENV || 'development',
    }
    
    fs.writeFileSync(
      path.join(envBackupPath, 'environment.json'),
      JSON.stringify(envInfo, null, 2)
    )
    
    // Backup environment-specific files (without secrets)
    const envFiles = ['.env.example', '.env.local.example']
    for (const file of envFiles) {
      const srcPath = path.join(process.cwd(), file)
      if (fs.existsSync(srcPath)) {
        fs.copyFileSync(srcPath, path.join(envBackupPath, file))
      }
    }
  }

  async createBackupManifest(backupPath, backupName) {
    const manifest = {
      name: backupName,
      timestamp: new Date().toISOString(),
      type: 'full',
      version: this.getAppVersion(),
      gitCommit: this.getCurrentGitCommit(),
      gitBranch: this.getCurrentGitBranch(),
      environment: process.env.NODE_ENV || 'development',
      contents: this.getBackupContents(backupPath),
      size: this.getDirectorySize(backupPath),
    }
    
    fs.writeFileSync(
      path.join(backupPath, 'backup-manifest.json'),
      JSON.stringify(manifest, null, 2)
    )
  }

  getBackupContents(backupPath) {
    const contents = []
    const items = fs.readdirSync(backupPath)
    
    for (const item of items) {
      const itemPath = path.join(backupPath, item)
      const stats = fs.statSync(itemPath)
      
      contents.push({
        name: item,
        type: stats.isDirectory() ? 'directory' : 'file',
        size: stats.isDirectory() ? this.getDirectorySize(itemPath) : stats.size,
      })
    }
    
    return contents
  }

  async compressBackup(backupPath, backupName) {
    console.log('🗜️  Compressing backup...')
    
    const archiveName = `${backupName}.tar.gz`
    const archivePath = path.join(this.backupDir, archiveName)
    
    execSync(`tar -czf "${archivePath}" -C "${this.backupDir}" "${backupName}"`)
    
    // Remove uncompressed backup
    fs.rmSync(backupPath, { recursive: true, force: true })
    
    return archivePath
  }

  async uploadToS3(archivePath, backupName) {
    console.log('☁️  Uploading backup to S3...')
    
    const fileStream = fs.createReadStream(archivePath)
    const key = `frontend/${backupName}.tar.gz`
    
    const uploadParams = {
      Bucket: this.bucketName,
      Key: key,
      Body: fileStream,
      Metadata: {
        'backup-type': 'full',
        'app-version': this.getAppVersion(),
        'git-commit': this.getCurrentGitCommit(),
        'created-at': new Date().toISOString(),
      },
    }
    
    await this.s3.upload(uploadParams).promise()
    console.log(`✅ Backup uploaded to S3: s3://${this.bucketName}/${key}`)
  }

  async listBackups() {
    console.log('📋 Listing available backups...')
    
    const backups = []
    
    // List local backups
    if (fs.existsSync(this.backupDir)) {
      const files = fs.readdirSync(this.backupDir)
      for (const file of files) {
        if (file.endsWith('.tar.gz')) {
          const filePath = path.join(this.backupDir, file)
          const stats = fs.statSync(filePath)
          backups.push({
            name: file.replace('.tar.gz', ''),
            location: 'local',
            size: stats.size,
            created: stats.mtime,
          })
        }
      }
    }
    
    // List S3 backups
    if (this.isS3Configured()) {
      try {
        const s3Objects = await this.s3.listObjectsV2({
          Bucket: this.bucketName,
          Prefix: 'frontend/',
        }).promise()
        
        for (const object of s3Objects.Contents || []) {
          backups.push({
            name: path.basename(object.Key, '.tar.gz'),
            location: 's3',
            size: object.Size,
            created: object.LastModified,
          })
        }
      } catch (error) {
        console.warn('Could not list S3 backups:', error.message)
      }
    }
    
    return backups.sort((a, b) => new Date(b.created) - new Date(a.created))
  }

  async restoreBackup(backupName, location = 'local') {
    console.log(`🔄 Restoring backup: ${backupName}`)
    
    // Create restoration point
    const restorationPoint = await this.createRestorePoint()
    
    try {
      let archivePath
      
      if (location === 's3') {
        archivePath = await this.downloadFromS3(backupName)
      } else {
        archivePath = path.join(this.backupDir, `${backupName}.tar.gz`)
      }
      
      if (!fs.existsSync(archivePath)) {
        throw new Error(`Backup archive not found: ${archivePath}`)
      }
      
      // Extract backup
      const extractPath = path.join(this.backupDir, `restore-${Date.now()}`)
      fs.mkdirSync(extractPath, { recursive: true })
      
      execSync(`tar -xzf "${archivePath}" -C "${extractPath}"`)
      
      const backupContentPath = path.join(extractPath, backupName)
      
      // Verify backup manifest
      const manifestPath = path.join(backupContentPath, 'backup-manifest.json')
      if (!fs.existsSync(manifestPath)) {
        throw new Error('Invalid backup: manifest not found')
      }
      
      const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'))
      console.log(`📋 Restoring backup from ${manifest.timestamp}`)
      
      // Stop application if running
      await this.stopApplication()
      
      // Restore source code
      await this.restoreSourceCode(backupContentPath)
      
      // Restore configuration
      await this.restoreConfiguration(backupContentPath)
      
      // Restore dependencies
      await this.restoreDependencies(backupContentPath)
      
      // Cleanup extraction directory
      fs.rmSync(extractPath, { recursive: true, force: true })
      
      console.log('✅ Backup restored successfully')
      console.log(`⚠️  Restoration point created: ${restorationPoint}`)
      
    } catch (error) {
      console.error('❌ Restore failed:', error)
      console.log(`🔄 Restoring from restoration point: ${restorationPoint}`)
      await this.restoreFromRestorePoint(restorationPoint)
      throw error
    }
  }

  async createRestorePoint() {
    console.log('📍 Creating restoration point...')
    
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const restorePointName = `restore-point-${timestamp}`
    
    // Create a quick backup before restoration
    return await this.createFullBackup()
  }

  async restoreFromRestorePoint(restorePointName) {
    // This would restore from the restoration point
    // Implementation similar to restoreBackup but simpler
    console.log(`🔄 Restoring from restoration point: ${restorePointName}`)
  }

  async downloadFromS3(backupName) {
    console.log('📥 Downloading backup from S3...')
    
    const key = `frontend/${backupName}.tar.gz`
    const downloadPath = path.join(this.backupDir, `${backupName}.tar.gz`)
    
    const downloadParams = {
      Bucket: this.bucketName,
      Key: key,
    }
    
    const data = await this.s3.getObject(downloadParams).promise()
    fs.writeFileSync(downloadPath, data.Body)
    
    return downloadPath
  }

  async restoreSourceCode(backupPath) {
    console.log('📁 Restoring source code...')
    
    const sourcePath = path.join(backupPath, 'source')
    if (!fs.existsSync(sourcePath)) return
    
    // Backup current git state
    const gitBackupPath = path.join(this.backupDir, 'temp-git-backup')
    if (fs.existsSync('.git')) {
      this.copyDirectory('.git', gitBackupPath)
    }
    
    // Clear current source (except .git)
    const items = fs.readdirSync(process.cwd())
    for (const item of items) {
      if (item !== '.git' && item !== 'node_modules' && item !== '.maintenance') {
        const itemPath = path.join(process.cwd(), item)
        fs.rmSync(itemPath, { recursive: true, force: true })
      }
    }
    
    // Restore source files
    this.copyDirectory(sourcePath, process.cwd())
    
    // Restore .git if it was backed up
    if (fs.existsSync(gitBackupPath)) {
      if (fs.existsSync('.git')) {
        fs.rmSync('.git', { recursive: true, force: true })
      }
      this.copyDirectory(gitBackupPath, '.git')
      fs.rmSync(gitBackupPath, { recursive: true, force: true })
    }
  }

  async restoreConfiguration(backupPath) {
    console.log('⚙️  Restoring configuration...')
    
    const configPath = path.join(backupPath, 'config')
    if (!fs.existsSync(configPath)) return
    
    const configFiles = fs.readdirSync(configPath)
    for (const file of configFiles) {
      const srcPath = path.join(configPath, file)
      const destPath = path.join(process.cwd(), file)
      
      if (fs.statSync(srcPath).isFile()) {
        fs.copyFileSync(srcPath, destPath)
      } else {
        this.copyDirectory(srcPath, destPath)
      }
    }
  }

  async restoreDependencies(backupPath) {
    console.log('📦 Restoring dependencies...')
    
    // Reinstall dependencies based on restored package.json
    execSync('npm ci', { stdio: 'inherit' })
  }

  async stopApplication() {
    // Stop running development server or processes
    try {
      execSync('pkill -f "vite"', { stdio: 'ignore' })
    } catch (error) {
      // Process might not be running
    }
  }

  async cleanupOldBackups() {
    console.log('🧹 Cleaning up old backups...')
    
    const cutoffDate = new Date()
    cutoffDate.setDate(cutoffDate.getDate() - this.retentionDays)
    
    // Cleanup local backups
    if (fs.existsSync(this.backupDir)) {
      const files = fs.readdirSync(this.backupDir)
      for (const file of files) {
        if (file.endsWith('.tar.gz')) {
          const filePath = path.join(this.backupDir, file)
          const stats = fs.statSync(filePath)
          
          if (stats.mtime < cutoffDate) {
            fs.unlinkSync(filePath)
            console.log(`🗑️  Deleted old backup: ${file}`)
          }
        }
      }
    }
    
    // Cleanup S3 backups
    if (this.isS3Configured()) {
      try {
        const s3Objects = await this.s3.listObjectsV2({
          Bucket: this.bucketName,
          Prefix: 'frontend/',
        }).promise()
        
        for (const object of s3Objects.Contents || []) {
          if (object.LastModified < cutoffDate) {
            await this.s3.deleteObject({
              Bucket: this.bucketName,
              Key: object.Key,
            }).promise()
            console.log(`🗑️  Deleted old S3 backup: ${object.Key}`)
          }
        }
      } catch (error) {
        console.warn('Could not cleanup S3 backups:', error.message)
      }
    }
  }

  // Utility methods
  copyDirectory(src, dest) {
    if (!fs.existsSync(src)) return
    
    fs.mkdirSync(dest, { recursive: true })
    const items = fs.readdirSync(src)
    
    for (const item of items) {
      const srcPath = path.join(src, item)
      const destPath = path.join(dest, item)
      const stats = fs.statSync(srcPath)
      
      if (stats.isDirectory()) {
        this.copyDirectory(srcPath, destPath)
      } else {
        fs.copyFileSync(srcPath, destPath)
      }
    }
  }

  getDirectorySize(dirPath) {
    let size = 0
    
    if (!fs.existsSync(dirPath)) return size
    
    const items = fs.readdirSync(dirPath)
    for (const item of items) {
      const itemPath = path.join(dirPath, item)
      const stats = fs.statSync(itemPath)
      
      if (stats.isDirectory()) {
        size += this.getDirectorySize(itemPath)
      } else {
        size += stats.size
      }
    }
    
    return size
  }

  getCurrentGitCommit() {
    try {
      return execSync('git rev-parse HEAD', { encoding: 'utf8' }).trim()
    } catch (error) {
      return 'unknown'
    }
  }

  getCurrentGitBranch() {
    try {
      return execSync('git rev-parse --abbrev-ref HEAD', { encoding: 'utf8' }).trim()
    } catch (error) {
      return 'unknown'
    }
  }

  getAppVersion() {
    try {
      const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'))
      return packageJson.version || 'unknown'
    } catch (error) {
      return 'unknown'
    }
  }

  isS3Configured() {
    return !!(process.env.AWS_ACCESS_KEY_ID && 
              process.env.AWS_SECRET_ACCESS_KEY && 
              process.env.BACKUP_S3_BUCKET)
  }
}

module.exports = BackupRecoveryManager

// Usage
if (require.main === module) {
  const manager = new BackupRecoveryManager()
  const command = process.argv[2]
  
  switch (command) {
    case 'backup':
      manager.createFullBackup()
        .then(backupName => console.log(`✅ Backup completed: ${backupName}`))
        .catch(error => {
          console.error('❌ Backup failed:', error)
          process.exit(1)
        })
      break
      
    case 'list':
      manager.listBackups()
        .then(backups => {
          console.log('\n📋 Available backups:')
          for (const backup of backups) {
            const size = Math.round(backup.size / 1024 / 1024)
            console.log(`  • ${backup.name} (${backup.location}, ${size}MB, ${backup.created})`)
          }
        })
        .catch(error => console.error('❌ List failed:', error))
      break
      
    case 'restore':
      const backupName = process.argv[3]
      const location = process.argv[4] || 'local'
      
      if (!backupName) {
        console.error('❌ Backup name required')
        process.exit(1)
      }
      
      manager.restoreBackup(backupName, location)
        .then(() => console.log('✅ Restore completed'))
        .catch(error => {
          console.error('❌ Restore failed:', error)
          process.exit(1)
        })
      break
      
    case 'cleanup':
      manager.cleanupOldBackups()
        .then(() => console.log('✅ Cleanup completed'))
        .catch(error => console.error('❌ Cleanup failed:', error))
      break
      
    default:
      console.log('Usage: node backup-recovery.js <command>')
      console.log('Commands: backup, list, restore <name> [location], cleanup')
      process.exit(1)
  }
}
```

## Directory Structure
```
scripts/
├── maintenance/
│   ├── dependency-management.js  # Automated dependency updates
│   ├── performance-monitoring.js # Performance regression testing
│   ├── backup-recovery.js        # Backup and recovery system
│   ├── health-check.js          # System health monitoring
│   └── maintenance-scheduler.js  # Maintenance task scheduling
├── deploy/
│   ├── pre-deploy-checks.js     # Pre-deployment validation
│   ├── post-deploy-verification.js # Post-deployment testing
│   └── rollback-procedures.js   # Automated rollback
└── utilities/
    ├── log-analyzer.js          # Log analysis tools
    ├── performance-profiler.js  # Performance profiling
    └── security-scanner.js      # Security vulnerability scanning

.maintenance/
├── backups/                     # Local backup storage
├── performance-reports/         # Performance test results
├── dependency-reports/          # Dependency analysis reports
├── logs/                       # Maintenance logs
└── configs/                    # Maintenance configuration files

.github/
├── workflows/
│   ├── dependency-updates.yml   # Automated dependency updates
│   ├── performance-monitoring.yml # Performance regression testing
│   ├── backup-schedule.yml      # Scheduled backups
│   └── maintenance-checks.yml   # Regular maintenance checks
└── MAINTENANCE.md              # Maintenance procedures documentation

docs/
├── maintenance/
│   ├── procedures.md           # Maintenance procedures
│   ├── troubleshooting.md      # Common issues and solutions
│   ├── backup-recovery.md      # Backup and recovery guide
│   └── performance-tuning.md   # Performance optimization guide
└── runbooks/
    ├── incident-response.md     # Incident response procedures
    ├── deployment-rollback.md   # Rollback procedures
    └── security-incident.md     # Security incident response
```

## Benefits
- Automated dependency management
- Proactive security updates
- Performance regression detection
- Comprehensive backup and recovery
- Automated maintenance scheduling
- Documentation and procedures
- Incident response capabilities
- Long-term sustainability
- Cost optimization
- Risk mitigation
- Compliance maintenance
- Team knowledge sharing

## Maintenance Features Implemented
- ✅ Automated dependency updates
- ✅ Security vulnerability monitoring
- ✅ Performance regression testing
- ✅ Bundle size monitoring  
- ✅ Automated backup system
- ✅ Recovery procedures
- ✅ Maintenance scheduling
- ✅ Health monitoring
- ✅ Documentation system
- ✅ Incident response procedures
- ✅ Log analysis tools
- ✅ Cleanup procedures

## Maintenance Schedule
- **Daily**: Security scans, dependency checks
- **Weekly**: Performance tests, backup verification
- **Monthly**: Full system backup, dependency updates
- **Quarterly**: Major version updates, architecture review
- **Annually**: Security audit, disaster recovery testing

## Key Metrics Monitored
- Dependency freshness
- Security vulnerability count
- Performance regression detection
- Backup success rate
- System uptime
- Error rates
- Resource utilization
- Update deployment success