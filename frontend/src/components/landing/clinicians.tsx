'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Stethoscope, ClipboardCheck, MessageSquare, TrendingUp } from 'lucide-react';
import Link from 'next/link';

const clinicianBenefits = [
  {
    icon: Stethoscope,
    title: 'Clinical Decision Support',
    description: 'AI-powered recommendations based on the latest medical research and patient data.',
  },
  {
    icon: ClipboardCheck,
    title: 'Streamlined Workflows',
    description: 'Reduce administrative burden and focus more time on patient care.',
  },
  {
    icon: MessageSquare,
    title: 'Intelligent Documentation',
    description: 'Automatically generate clinical notes and documentation with AI assistance.',
  },
  {
    icon: TrendingUp,
    title: 'Data-Driven Insights',
    description: 'Track patient outcomes and identify trends to improve care quality.',
  },
];

export default function Clinicians() {
  return (
    <section className="py-24 bg-background" id="clinicians">
      <div className="container mx-auto px-6">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Left: Content */}
          <div>
            <h2 className="text-4xl md:text-5xl font-bold text-foreground mb-6">
              Built for Clinicians, By Healthcare Professionals
            </h2>
            <p className="text-lg text-muted-foreground mb-8">
              HealthSecure is designed to enhance your clinical practice with intelligent tools that save time, improve accuracy, and support better patient outcomes.
            </p>
            
            <div className="space-y-6 mb-8">
              {clinicianBenefits.map((benefit, index) => {
                const Icon = benefit.icon;
                return (
                  <div key={index} className="flex items-start space-x-4">
                    <div className="flex-shrink-0">
                      <div className="w-12 h-12 bg-secondary/10 rounded-lg flex items-center justify-center">
                        <Icon className="w-6 h-6 text-secondary" />
                      </div>
                    </div>
                    <div>
                      <h3 className="font-semibold text-foreground mb-1">{benefit.title}</h3>
                      <p className="text-muted-foreground">{benefit.description}</p>
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="flex flex-wrap gap-4">
              <Link href="/demo">
                <Button size="lg" className="font-semibold">
                  Schedule a Demo
                </Button>
              </Link>
              <Link href="/case-studies">
                <Button variant="outline" size="lg" className="font-semibold">
                  View Case Studies
                </Button>
              </Link>
            </div>
          </div>

          {/* Right: Visual */}
          <div className="relative">
            <Card className="border-2 shadow-xl">
              <CardContent className="p-8">
                <div className="aspect-square bg-gradient-to-br from-primary/20 via-secondary/20 to-accent/20 rounded-lg flex items-center justify-center">
                  <div className="text-center">
                    <Stethoscope className="w-24 h-24 text-primary mx-auto mb-4" />
                    <p className="text-2xl font-semibold text-foreground">
                      Trusted by Healthcare Professionals
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </section>
  );
}
