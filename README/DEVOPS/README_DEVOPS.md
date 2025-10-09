# 🧠 DevOps & CI/CD Overview — SocialPredict

This document provides a high-level overview of the **DevOps architecture** and **continuous integration / continuous deployment (CI/CD)** practices for the **SocialPredict** project.  
It includes a detailed explanation of how the **GitHub Actions pipeline** is structured, how Go version testing is automated, and how branch protection integrates with CI to maintain code quality.

---

## ⚙️ Overview of the DevOps System

The SocialPredict backend uses **GitHub Actions** for continuous integration and code-quality enforcement.  
All build, test, and analysis pipelines run automatically when pull requests are opened or updated.

### Goals of the DevOps System
- **Consistency:** Ensure every PR runs the same validation steps before merging.
- **Security:** Block merges until required checks pass.
- **Scalability:** Support multiple Go versions and environments without manual updates.
- **Simplicity:** Use stable CI job names so branch protection never breaks.

---

## 🏗️ CI/CD Pipeline Architecture

The current CI/CD pipeline is defined in **`.github/workflows/backend.yml`** and built around modular jobs that can evolve over time.  

### Pipeline Jobs

#### 1. `prepare-matrix`
This job dynamically generates the list of Go versions to test against.  
It reads from the `.github/go-versions` file (if present), or defaults to the `go` version declared in `backend/go.mod`.

**Purpose:**
- Build a JSON matrix of Go versions at runtime.
- Allow flexible control over which Go versions are tested.
- Avoid hard-coding versions directly in the workflow.

---

#### 2. `smoke`
A fast-running integration check to verify that the backend can build and start successfully.  
Currently pinned to Go `1.25.x`, but can be converted to the dynamic pattern later.

**Purpose:**
- Catch obvious build or runtime issues quickly.
- Provide early failure before running the full test suite.

---

#### 3. `unit_matrix`
This job runs the full Go test suite for each Go version defined by `prepare-matrix`.  
It uses the matrix output to automatically handle multiple Go versions in parallel.

**Key Features:**
- Uses `actions/setup-go@v5`
- `check-latest: true` ensures the newest patch release
- `cache: true` enables Go module caching for faster runs
- Depends on the successful completion of `smoke` and `prepare-matrix`

---

#### 4. `unit` (Aggregator)
The aggregator job reports the **final unified test result** back to GitHub.  
It depends on all matrix jobs and produces a **single, stable status check name** for branch protection.

**Purpose:**
- Simplify merge gating — only one required check ever.
- Ensure consistent status reporting regardless of Go versions tested.

---

## 🧩 Dynamic Go Version Matrix System

The dynamic matrix system allows SocialPredict to test across any number of Go versions without editing the workflow file.

### How It Works
- `.github/go-versions` lists which versions to test.
- If this file is missing, the pipeline automatically detects the version from `backend/go.mod`.
- Both explicit (`1.26.x`) and file-based (`file:backend/go.mod`) entries are supported.

### Example Configurations

```
To test only your current module version:

file:backend/go.mod


To test multiple Go versions:

1.25.x
1.26.x


To combine both:

file:backend/go.mod
1.25.x
1.26.x
```


This approach allows developers to add or remove Go versions by editing one file — no YAML modification or branch-protection changes needed.

---

## 🧱 Branch Protection Setup

### Why It Matters
GitHub branch protection enforces that key checks pass before merging.  
Previously, changing Go versions (like from `1.23.x` to `1.25.x`) broke branch protection because the check name changed.  
Now, with a stable aggregator job named `unit`, protection remains valid forever.

### Steps to Configure

1. Navigate to your repository:  
   **Settings → Branches → Branch protection rule for `main`**
2. Under **Require status checks to pass before merging**, configure:
   - ✅ **Add:** `Backend / unit (pull_request)`
   - 🗑️ **Remove:** versioned checks like  
     `unit (1.23.x)`  
     `unit (1.25.x)`  
     `unit (matrix) (file, backend/go.mod)`
3. Keep other checks (like CodeQL or smoke) if desired.
4. Save.

From this point on, only `Backend / unit (pull_request)` needs to pass for merging.

---

## 🪄 Benefits of This Implementation

### ✅ No More Branch Protection Breakage
Branch protection now keys off the stable `unit` check name — version bumps no longer require any GitHub settings changes.

### ✅ Flexible Version Management
You can modify `.github/go-versions` freely to test any Go versions without editing YAML.

### ✅ Fast & Cached Builds
Replaced the deprecated string cache option (`'gomod'`) with a boolean (`true`), enabling caching without YAML validation errors.

### ✅ Cleaner CI Signal
All Go versions funnel through a single pass/fail check, so reviewers see one unified result instead of multiple scattered ones.

### ✅ Extensible for Future Jobs
The `prepare-matrix` pattern can later power:
- Dynamic `lint`, `build`, or `security-scan` jobs
- Environment-specific deploy pipelines

---

## 🧰 File Summary

| File | Purpose |
|------|----------|
| `.github/workflows/backend.yml` | Main GitHub Actions pipeline definition |
| `.github/go-versions` | Defines Go versions to test |
| `backend/go.mod` | Fallback source for default Go version |
| `README_DEVOPS.md` | This documentation |

---

## 🔧 Fixes Applied

### YAML Cache Parameter Fix
**Issue:**  
`actions/setup-go@v5` failed with:  
_Input does not meet YAML 1.2 "Core Schema" specification: cache_

**Cause:**  
Used `cache: 'gomod'` (string), but v5 only supports boolean values.

**Fix:**  
Changed to `cache: true` in both setup steps of the `unit_matrix` job.

**Result:**  
The pipeline now correctly caches dependencies and passes validation.

---

## 🧭 Future Recommendations
- Convert `smoke` to dynamic matrix structure.
- Add linting and build-verification stages.
- Extend matrix-driven testing to front-end workflows if needed.
- Optionally integrate container image build and deployment (e.g., via GitHub Environments or Docker Hub actions).

---

## ✅ Summary

| Feature | Benefit |
|----------|----------|
| Dynamic version matrix | Automatically follows `.github/go-versions` or `go.mod` |
| Stable `unit` aggregator | Single permanent required check |
| Simplified branch protection | No stale “Expected” checks |
| YAML cache fix | Faster, reliable builds |
| Extensible pattern | Reusable for other job types |

---

**In short:**  
SocialPredict’s CI/CD pipeline now supports dynamic Go version testing, stable branch protection, and modern GitHub Actions conventions — all while reducing maintenance overhead and ensuring future scalability.
