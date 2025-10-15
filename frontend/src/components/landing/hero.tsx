'use client';

import { useEffect } from 'react';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { ArrowRight } from 'lucide-react';

export default function Hero() {
  useEffect(() => {
    // Load Unicorn Studio script
    if (!window.UnicornStudio) {
      window.UnicornStudio = { isInitialized: false };
      const script = document.createElement('script');
      script.src = 'https://cdn.jsdelivr.net/gh/hiunicornstudio/unicornstudio.js@v1.4.33/dist/unicornStudio.umd.js';
      script.onload = () => {
        if (!window.UnicornStudio.isInitialized) {
          window.UnicornStudio.init();
          window.UnicornStudio.isInitialized = true;
        }
      };
      (document.head || document.body).appendChild(script);
    }
  }, []);

  return (
    <section className="relative min-h-screen flex items-end pb-32" id="home">
      {/* Interactive Background - Unicorn Studio */}
      <div
        id="unicorn-background"
        className="absolute inset-0 -z-10 overflow-hidden"
        style={{
          background: 'linear-gradient(135deg, oklch(0.9789 0.0082 121.6272) 0%, oklch(0.95 0.02 200) 100%)',
          willChange: 'transform',
        }}
      >
        {/* Unicorn Studio Animation */}
        <div 
          className="absolute inset-0 flex items-center justify-center"
          style={{
            transform: 'scale(1.1)',
            transformOrigin: 'center center',
            willChange: 'transform',
          }}
        >
          <div 
            data-us-project="KKGaPr6wLLMDRDcueNOw" 
            style={{ 
              width: '100%', 
              height: '100%',
              minWidth: '1440px',
              minHeight: '900px',
              willChange: 'transform',
            }}
          />
        </div>

        {/* Gradient Overlay for better text readability */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/40 via-black/20 to-transparent" />
      </div>

      {/* Hero Content */}
      <div className="container mx-auto px-6 relative z-10">
        <div className="max-w-3xl space-y-6">
          <h1 className="text-5xl md:text-7xl font-bold text-white leading-tight drop-shadow-2xl">
            Secure, Intelligent Healthcare is Here.
          </h1>
          <p className="text-xl md:text-2xl text-white/95 max-w-2xl drop-shadow-xl">
            AI-powered insights, HIPAA-compliant security. Empower your practice with HealthSecure.
          </p>
          <div className="pt-4">
            <Link href="#features">
              <Button size="lg" className="text-lg px-8 py-6 group shadow-2xl">
                Explore the Platform
                <ArrowRight className="ml-2 h-5 w-5 transition-transform group-hover:translate-x-1" />
              </Button>
            </Link>
          </div>
        </div>
      </div>

      {/* Scroll Indicator */}
      <div className="absolute bottom-8 left-1/2 transform -translate-x-1/2 animate-bounce z-10">
        <div className="w-6 h-10 border-2 border-white/40 rounded-full flex items-start justify-center p-2">
          <div className="w-1 h-3 bg-white/40 rounded-full"></div>
        </div>
      </div>
    </section>
  );
}
