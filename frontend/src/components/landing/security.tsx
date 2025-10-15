'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Shield, Lock, Eye, FileCheck, Server, AlertTriangle } from 'lucide-react';
import Link from 'next/link';

const securityFeatures = [
  {
    icon: Shield,
    title: 'HIPAA Compliant',
    description: 'Full compliance with HIPAA regulations for healthcare data protection.',
  },
  {
    icon: Lock,
    title: 'End-to-End Encryption',
    description: 'Advanced encryption protocols protect data in transit and at rest.',
  },
  {
    icon: Eye,
    title: 'Access Controls',
    description: 'Role-based access control ensures only authorized personnel can view sensitive data.',
  },
  {
    icon: FileCheck,
    title: 'Audit Logging',
    description: 'Comprehensive audit trails track every action for complete accountability.',
  },
  {
    icon: Server,
    title: 'Secure Infrastructure',
    description: 'Enterprise-grade servers with redundancy and disaster recovery protocols.',
  },
  {
    icon: AlertTriangle,
    title: 'Emergency Protocols',
    description: 'Secure emergency access with full audit trails for critical situations.',
  },
];

export default function Security() {
  return (
    <section className="py-24 bg-muted/30" id="security">
      <div className="container mx-auto px-6">
        <div className="text-center mb-16">
          <div className="inline-block p-3 bg-primary/10 rounded-full mb-4">
            <Shield className="w-12 h-12 text-primary" />
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-foreground mb-4">
            Security You Can Trust
          </h2>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            Your patients' data deserves the highest level of protection. HealthSecure delivers enterprise-grade security with HIPAA compliance built-in.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
          {securityFeatures.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <Card key={index} className="border bg-card hover:shadow-md transition-shadow">
                <CardContent className="pt-6">
                  <div className="flex items-start space-x-4">
                    <div className="flex-shrink-0">
                      <div className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                        <Icon className="w-5 h-5 text-primary" />
                      </div>
                    </div>
                    <div>
                      <h3 className="font-semibold text-foreground mb-2">{feature.title}</h3>
                      <p className="text-sm text-muted-foreground">{feature.description}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>

        <div className="text-center">
          <Link href="/security">
            <Button variant="outline" size="lg" className="font-semibold">
              Learn More About Our Security
            </Button>
          </Link>
        </div>
      </div>
    </section>
  );
}
