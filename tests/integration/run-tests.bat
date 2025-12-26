@echo off
REM
REM GIIA Integration Test Suite Runner (Windows)
REM
REM This script starts the test environment, waits for services,
REM runs integration tests, and cleans up.
REM
REM Usage:
REM   run-tests.bat              - Run all tests
REM   run-tests.bat -v           - Run with verbose output
REM   run-tests.bat -run "Auth"  - Run specific test pattern
REM

setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

REM Default values
set "VERBOSE="
set "TEST_PATTERN="
set "TIMEOUT=10m"
set "SKIP_SETUP=false"
set "SKIP_TEARDOWN=false"

REM Parse arguments
:parse_args
if "%~1"=="" goto :end_parse
if /i "%~1"=="-v" (
    set "VERBOSE=-v"
    shift
    goto :parse_args
)
if /i "%~1"=="--verbose" (
    set "VERBOSE=-v"
    shift
    goto :parse_args
)
if /i "%~1"=="-run" (
    set "TEST_PATTERN=-run %~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="--timeout" (
    set "TIMEOUT=%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="--skip-setup" (
    set "SKIP_SETUP=true"
    shift
    goto :parse_args
)
if /i "%~1"=="--skip-teardown" (
    set "SKIP_TEARDOWN=true"
    shift
    goto :parse_args
)
if /i "%~1"=="-h" goto :show_help
if /i "%~1"=="--help" goto :show_help
echo Unknown option: %~1
exit /b 1

:show_help
echo Usage: %~nx0 [options]
echo.
echo Options:
echo   -v, --verbose       Run tests with verbose output
echo   -run PATTERN        Run only tests matching pattern
echo   --timeout DURATION  Set test timeout (default: 10m)
echo   --skip-setup        Skip docker-compose up
echo   --skip-teardown     Skip docker-compose down
echo   -h, --help          Show this help message
exit /b 0

:end_parse

echo ================================================================
echo      GIIA Integration Test Suite
echo ================================================================
echo.

REM Start test environment
if "%SKIP_SETUP%"=="false" (
    echo [*] Starting test environment...
    docker-compose -f docker-compose.yml up -d
    
    echo [*] Waiting for services to be healthy...
    
    REM Wait for services (simplified for Windows)
    timeout /t 30 /nobreak > nul
    
    echo [OK] Services should be ready
    echo.
) else (
    echo [*] Skipping setup (--skip-setup)
)

REM Run integration tests
echo [*] Running integration tests...
echo.

REM Build test command
set "TEST_CMD=go test %VERBOSE% -timeout %TIMEOUT% %TEST_PATTERN% ./..."
echo Command: %TEST_CMD%
echo.

REM Run tests
%TEST_CMD%
set "TEST_EXIT_CODE=%errorlevel%"

echo.

REM Show results
if %TEST_EXIT_CODE%==0 (
    echo ================================================================
    echo      [OK] All tests passed!
    echo ================================================================
) else (
    echo ================================================================
    echo      [FAIL] Some tests failed!
    echo ================================================================
)

REM Cleanup
if "%SKIP_TEARDOWN%"=="false" (
    echo.
    echo [*] Cleaning up test environment...
    docker-compose -f docker-compose.yml down -v
    echo [OK] Cleanup complete!
) else (
    echo [*] Skipping teardown (--skip-teardown)
)

exit /b %TEST_EXIT_CODE%
