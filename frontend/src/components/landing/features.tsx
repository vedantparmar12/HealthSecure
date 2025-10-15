'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Brain, Shield, Users, FileText, Clock, Activity } from 'lucide-react';

const features = [
  {
    icon: Brain,
    title: 'AI-Powered Knowledge Base',
    description: 'Instantly access and explore medical knowledge with our intelligent AI assistant.',
  },
  {
    icon: Shield,
    title: 'HIPAA-Compliant Security',
    description: 'Rest easy knowing all your data is protected with the highest security standards.',
  },
  {
    icon: Users,
    title: 'Patient Management',
    description: 'Streamline patient records, medical history, and treatment plans in one secure location.',
  },
  {
    icon: FileText,
    title: 'Comprehensive Audit Trails',
    description: 'Maintain compliance with detailed logs of all system access and modifications.',
  },
  {
    icon: Clock,
    title: 'Real-Time Insights',
    description: 'Get instant AI-powered analysis and recommendations for better patient care.',
  },
  {
    icon: Activity,
    title: 'Emergency Access',
    description: 'Secure emergency access protocols ensure critical care is never delayed.',
  },
];

export default function Features() {
  return (
    <section className="py-24 bg-background" id="features">
      <div className="container mx-auto px-6">
        <div className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl font-bold text-foreground mb-4">
            Everything You Need for Modern Healthcare
          </h2>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            Powerful features designed to enhance patient care while maintaining the highest security standards.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <Card
                key={index}
                className="border-2 hover:border-primary transition-all duration-300 hover:shadow-lg"
              >
                <CardHeader>
                  <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
                    <Icon className="w-6 h-6 text-primary" />
                  </div>
                  <CardTitle className="text-xl">{feature.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-base">{feature.description}</CardDescription>
                </CardContent>
              </Card>
            );
          })}
        </div>
      </div>
    </section>
  );
}
