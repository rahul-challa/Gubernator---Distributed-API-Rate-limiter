#!/usr/bin/env python3
"""
Gubernator Rate Limiter Visualization Script
Floods the API with requests and displays a real-time visualization
of rate limiting in action.
"""

import requests
import time
import threading
import sys
from collections import deque
from datetime import datetime

# Configuration
API_URL = "http://localhost:8080/api/v1/test"
REQUESTS_PER_SECOND = 50
DURATION_SECONDS = 30
MAX_BARS = 80  # Terminal width for visualization

class RateLimitVisualizer:
    def __init__(self, api_url, rps, duration):
        self.api_url = api_url
        self.rps = rps
        self.duration = duration
        self.results = deque(maxlen=1000)  # Store last 1000 results
        self.running = True
        self.lock = threading.Lock()
        
    def make_request(self):
        """Make a single HTTP request and record the result"""
        try:
            start = time.time()
            response = requests.get(self.api_url, timeout=1)
            elapsed = time.time() - start
            
            with self.lock:
                self.results.append({
                    'status': response.status_code,
                    'timestamp': time.time(),
                    'elapsed': elapsed,
                    'remaining': response.headers.get('X-RateLimit-Remaining', 'N/A'),
                })
            
            return response.status_code
        except Exception as e:
            with self.lock:
                self.results.append({
                    'status': 0,
                    'timestamp': time.time(),
                    'elapsed': 0,
                    'error': str(e),
                })
            return 0
    
    def request_worker(self):
        """Worker thread that makes requests at specified rate"""
        interval = 1.0 / self.rps
        end_time = time.time() + self.duration
        
        while self.running and time.time() < end_time:
            self.make_request()
            time.sleep(interval)
    
    def display_stats(self):
        """Display real-time statistics and visualization"""
        while self.running:
            time.sleep(0.5)  # Update every 500ms
            
            with self.lock:
                if not self.results:
                    continue
                
                recent = list(self.results)[-MAX_BARS:]
                total = len(recent)
                
                if total == 0:
                    continue
                
                # Count status codes
                status_200 = sum(1 for r in recent if r['status'] == 200)
                status_429 = sum(1 for r in recent if r['status'] == 429)
                status_other = total - status_200 - status_429
                
                # Calculate percentages
                pct_200 = (status_200 / total) * 100 if total > 0 else 0
                pct_429 = (status_429 / total) * 100 if total > 0 else 0
                
                # Build visualization bar
                bar_200 = int((status_200 / total) * MAX_BARS) if total > 0 else 0
                bar_429 = int((status_429 / total) * MAX_BARS) if total > 0 else 0
                
                # Clear screen and print stats
                sys.stdout.write('\033[2J\033[H')  # Clear screen
                sys.stdout.write("=" * 80 + "\n")
                sys.stdout.write("Gubernator Rate Limiter - Real-time Visualization\n")
                sys.stdout.write("=" * 80 + "\n")
                sys.stdout.write(f"API URL: {self.api_url}\n")
                sys.stdout.write(f"Request Rate: {self.rps} req/s\n")
                sys.stdout.write(f"Total Requests (last {total}): {len(self.results)}\n")
                sys.stdout.write("-" * 80 + "\n")
                sys.stdout.write(f"Status 200 (OK):        {status_200:4d} ({pct_200:5.1f}%) {'█' * bar_200}\n")
                sys.stdout.write(f"Status 429 (Blocked):   {status_429:4d} ({pct_429:5.1f}%) {'█' * bar_429}\n")
                if status_other > 0:
                    sys.stdout.write(f"Other Status Codes:     {status_other:4d}\n")
                sys.stdout.write("-" * 80 + "\n")
                
                # Show recent request timeline
                sys.stdout.write("Recent Request Timeline (last 80 requests):\n")
                timeline = ""
                for r in recent[-80:]:
                    if r['status'] == 200:
                        timeline += "█"  # Green (allowed)
                    elif r['status'] == 429:
                        timeline += "░"  # Red (blocked)
                    else:
                        timeline += "?"  # Error
                
                sys.stdout.write(timeline + "\n")
                sys.stdout.write("█ = Allowed (200)  ░ = Blocked (429)  ? = Error\n")
                sys.stdout.write("=" * 80 + "\n")
                sys.stdout.flush()
    
    def run(self):
        """Run the visualization"""
        print(f"Starting flood test: {self.rps} req/s for {self.duration} seconds")
        print(f"Target: {self.api_url}")
        print("Press Ctrl+C to stop early\n")
        time.sleep(2)
        
        # Start display thread
        display_thread = threading.Thread(target=self.display_stats, daemon=True)
        display_thread.start()
        
        # Start request workers
        num_workers = max(1, self.rps // 10)  # One worker per 10 req/s
        workers = []
        
        for i in range(num_workers):
            worker_rps = self.rps // num_workers
            if i == 0:
                worker_rps += self.rps % num_workers  # Add remainder to first worker
            
            worker = threading.Thread(target=self.request_worker, daemon=True)
            workers.append(worker)
            worker.start()
        
        # Wait for duration
        try:
            time.sleep(self.duration)
        except KeyboardInterrupt:
            print("\n\nStopping...")
        
        self.running = False
        time.sleep(1)  # Let threads finish
        
        # Final summary
        with self.lock:
            all_results = list(self.results)
            total = len(all_results)
            if total > 0:
                final_200 = sum(1 for r in all_results if r['status'] == 200)
                final_429 = sum(1 for r in all_results if r['status'] == 429)
                
                print("\n" + "=" * 80)
                print("Final Statistics:")
                print(f"Total Requests: {total}")
                print(f"Allowed (200):  {final_200} ({final_200/total*100:.1f}%)")
                print(f"Blocked (429):   {final_429} ({final_429/total*100:.1f}%)")
                print("=" * 80)

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Flood test for Gubernator rate limiter")
    parser.add_argument("--url", default=API_URL, help="API endpoint URL")
    parser.add_argument("--rps", type=int, default=REQUESTS_PER_SECOND, help="Requests per second")
    parser.add_argument("--duration", type=int, default=DURATION_SECONDS, help="Test duration in seconds")
    
    args = parser.parse_args()
    
    visualizer = RateLimitVisualizer(args.url, args.rps, args.duration)
    visualizer.run()

