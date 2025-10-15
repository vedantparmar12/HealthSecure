'use client';

import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import Link from 'next/link';
import { ArrowRight, Mail } from 'lucide-react';

export default function CTA() {
  return (
    <section className="py-24 bg-gradient-to-br from-primary via-primary to-secondary" id="contact">
      <div className="container mx-auto px-6">
        <Card className="border-0 shadow-2xl">
          <div className="p-12 md:p-16 text-center">
            <h2 className="text-4xl md:text-5xl font-bold text-foreground mb-6">
              Ready to Transform Your Practice?
            </h2>
            <p className="text-xl text-muted-foreground mb-10 max-w-3xl mx-auto">
              Join healthcare professionals who trust HealthSecure for secure, intelligent patient care. 
              Get started today with a personalized demo.
            </p>
            
            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
              <Link href="/demo">
                <Button size="lg" className="text-lg px-8 py-6 bg-primary hover:bg-primary/90 group">
                  Get Started
                  <ArrowRight className="ml-2 h-5 w-5 transition-transform group-hover:translate-x-1" />
                </Button>
              </Link>
              <Link href="mailto:contact@healthsecure.com">
                <Button variant="outline" size="lg" className="text-lg px-8 py-6">
                  <Mail className="mr-2 h-5 w-5" />
                  Contact Sales
                </Button>
              </Link>
            </div>

            <div className="mt-12 pt-8 border-t">
              <div className="grid md:grid-cols-3 gap-8 text-center">
                <div>
                  <p className="text-3xl font-bold text-primary mb-2">100%</p>
                  <p className="text-muted-foreground">HIPAA Compliant</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-primary mb-2">24/7</p>
                  <p className="text-muted-foreground">Support Available</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-primary mb-2">500+</p>
                  <p className="text-muted-foreground">Healthcare Providers</p>
                </div>
              </div>
            </div>
          </div>
        </Card>
      </div>
    </section>
  );
}
